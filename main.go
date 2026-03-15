package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"browser/css"
	"browser/dom"
	"browser/js"
	"browser/layout"
	"browser/render"
	"browser/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <url>")
		os.Exit(1)
	}

	startURL := os.Args[1]

	// Create browser window
	browser := render.NewBrowser(900, 600)

	// When link is clicked or Go pressed, load the page
	browser.OnNavigate = func(req render.NavigationRequest) {
		loadPage(browser, req)
	}
	// Load initial page
	loadPage(browser, render.NavigationRequest{
		URL:    startURL,
		Method: "GET",
	})

	// Run the GUI
	browser.Run()
}

func loadPage(browser *render.Browser, req render.NavigationRequest) {
	pageURL := req.URL
	method := req.Method
	if method == "" {
		method = "GET"
	}
	fmt.Printf("Fetching (%s): %s\n", method, pageURL)
	browser.ShowLoading()
	browser.UpdateURLBar(pageURL)

	// Run fetch in background so UI stays responsive
	go func() {
		resp, err := utils.DoRequest(utils.HTTPRequest{
			Method:         method,
			URL:            pageURL,
			Body:           req.Body,
			ContentType:    req.ContentType,
			FormData:       req.Data,
			ReferrerPolicy: req.ReferrerPolicy,
			FromURL:        browser.GetCurrentURL(),
		})

		if err != nil {
			fmt.Println("Error:", err)
			browser.ShowError("Error 404")
			return
		}
		defer resp.Body.Close()

		fmt.Println("Parsing HTML...")
		document := dom.Parse(resp.Body)
		if document == nil {
			browser.ShowError("Error 404")
			fmt.Println("Error: failed to parse HTML")
			return
		}

		title := dom.FindTitle(document)
		browser.SetTitle(title)
		browser.SetDocument(document)

		fmt.Println("Fetching CSS...")

		// 1. Fetch external stylesheets in parallel
		links := dom.FindStylesheetLinks(document)
		cssResults := make([]string, len(links))
		var wg sync.WaitGroup

		for i, link := range links {
			wg.Add(1)
			go func(idx int, href string) {
				defer wg.Done()
				absURL := resolveURL(pageURL, href)
				fmt.Println("Fetching CSS:", absURL)
				cssResp, err := http.Get(absURL)
				if err == nil {
					data, _ := io.ReadAll(cssResp.Body)
					cssResp.Body.Close()
					// Resolve @import directives in fetched stylesheet
					seen := map[string]bool{absURL: true}
					cssResults[idx] = resolveCSSimports(string(data), absURL, 0, seen)
				} else {
					fmt.Println("Failed to fetch CSS:", err)
				}
			}(i, link)
		}

		wg.Wait()

		// Combine external CSS in order
		var externalCSS strings.Builder
		for _, cssContent := range cssResults {
			externalCSS.WriteString(cssContent + "\n")
		}

		// Store external CSS for reflow (when styles are disabled/enabled)
		browser.SetExternalCSS(externalCSS.String())

		// Combine external + internal <style> content (resolve @imports in inline styles)
		fullCSS := combineCSS(externalCSS.String(), document, pageURL)

		fmt.Println("Building layout...")
		stylesheet := css.Parse(fullCSS)
		browser.SetDocument(document)
		matchCtx := css.MatchContext{
			IsVisited:  func(url string) bool { return browser.IsVisited(url) },
			ResolveURL: func(href string) string { return resolveURL(pageURL, href) },
		}
		layoutTree := layout.BuildLayoutTree(document, stylesheet, layout.Viewport{
			Width:  float64(browser.Width),
			Height: float64(browser.Height),
		}, matchCtx)
		layout.ComputeLayout(layoutTree, float64(browser.Width))

		// Execute JavaScript
		fmt.Println("Executing JavaScript...")
		jsRuntime := js.NewJSRuntime(document, func() {
			browser.Reflow(browser.Width)
		})

		jsRuntime.SetAlertHandler(browser.ShowAlert)
		jsRuntime.SetConfirmHandler(browser.ShowConfirm)
		jsRuntime.SetPromptHandler(browser.ShowPrompt)
		browser.SetJSClickHandler(jsRuntime.DispatchClick)
		browser.SetBeforeNavigateHandler(jsRuntime.CheckBeforeUnload)

		jsRuntime.SetCurrentURL(pageURL)

		scripts := js.FindScripts(document)
		for i, script := range scripts {
			fmt.Printf("Running script %d...\n", i+1)
			jsRuntime.Execute(script)
		}

		browser.SetCurrentURL(pageURL)
		jsRuntime.SetReloadHandler(func() {
			browser.Refresh()
		})

		jsRuntime.SetTitleChangeHandler(browser.SetTitle)

		// Re-parse CSS after JavaScript (respects disabled styles)
		fullCSS = combineCSS(externalCSS.String(), document, pageURL)
		stylesheet = css.Parse(fullCSS)

		// Rebuild layout tree AFTER JavaScript has modified the DOM
		layoutTree = layout.BuildLayoutTree(document, stylesheet, layout.Viewport{
			Width:  float64(browser.Width),
			Height: float64(browser.Height),
		}, matchCtx)
		layout.ComputeLayout(layoutTree, float64(browser.Width))
		browser.SetContent(layoutTree)

		fmt.Println("Firing load event...")
		jsRuntime.FireLoad()

		browser.AddToHistory(pageURL)
		browser.MarkVisited(pageURL)

		fmt.Println("Page loaded!")
	}()
}

// combineCSS merges external CSS with inline <style> content, resolving @imports in inline styles.
func combineCSS(externalCSS string, document *dom.Node, pageURL string) string {
	inlineCSS := resolveCSSimports(dom.FindActiveStyleContent(document), pageURL, 0, map[string]bool{})
	return externalCSS + inlineCSS
}

func resolveCSSimports(cssContent, baseURL string, depth int, seen map[string]bool) string {
	if depth >= 5 {
		return cssContent
	}

	sheet := css.Parse(cssContent)
	if len(sheet.Imports) == 0 {
		return cssContent
	}

	var imported strings.Builder
	for _, importURL := range sheet.Imports {
		absURL := resolveURL(baseURL, importURL)
		if seen[absURL] {
			fmt.Printf("Skipping circular @import: %s\n", absURL)
			continue
		}
		seen[absURL] = true

		fmt.Printf("Fetching @import: %s\n", absURL)
		resp, err := http.Get(absURL)
		if err != nil {
			fmt.Printf("Failed to fetch @import %s: %v\n", absURL, err)
			continue
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Recursively resolve nested imports
		resolved := resolveCSSimports(string(data), absURL, depth+1, seen)
		imported.WriteString(resolved)
		imported.WriteString("\n")
	}

	// Imported rules prepended = lower cascade priority
	return imported.String() + cssContent
}

func resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}
