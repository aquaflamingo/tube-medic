package handlers

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/aqfl/tmcore"
	"github.com/aqfl/tubemedic-web/templates"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(
		templates.FS,
		"layout.html",
		"index.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "layout", nil)
}

func HandleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(errorBox("Form Parse Error", err.Error())))
		return
	}

	channelURL := r.FormValue("channel")
	if channelURL == "" {
		w.Write([]byte(errorBox("Validation Error", "YouTube channel URL is required.")))
		return
	}

	apiKey := os.Getenv("YT_API_KEY")
	if apiKey == "" {
		w.Write([]byte(errorBox("Server Configuration Error", "YT_API_KEY is not configured on the server. Please add it to your .env file.")))
		return
	}

	maxVideos := 50
	if v := r.FormValue("max_videos"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			maxVideos = n
		}
	}

	report, err := tmcore.RunScan(tmcore.Config{
		APIKey:     apiKey,
		ChannelURL: channelURL,
		MaxVideos:  maxVideos,
	})
	if err != nil {
		w.Write([]byte(errorBox("Scan Failed", fmt.Sprintf("Failed to scan channel: %s", err.Error()))))
		return
	}

	writeReport(w, report)
}

func writeReport(w http.ResponseWriter, r *tmcore.Report) {
	s := r.Summary

	statCards := fmt.Sprintf(`
	<div class="bg-white p-4 rounded-xl border border-blue-100 shadow-sm text-center">
		<div class="text-2xl font-bold text-blue-600">%d</div>
		<div class="text-xs text-blue-800 font-semibold uppercase tracking-wider mt-1">Videos Scanned</div>
	</div>
	<div class="bg-white p-4 rounded-xl border border-gray-100 shadow-sm text-center">
		<div class="text-2xl font-bold text-gray-600">%d</div>
		<div class="text-xs text-gray-800 font-semibold uppercase tracking-wider mt-1">Total Links</div>
	</div>
	<div class="bg-white p-4 rounded-xl border border-green-100 shadow-sm text-center">
		<div class="text-2xl font-bold text-green-600">%d</div>
		<div class="text-xs text-green-800 font-semibold uppercase tracking-wider mt-1">Working</div>
	</div>
	<div class="bg-white p-4 rounded-xl border border-yellow-100 shadow-sm text-center">
		<div class="text-2xl font-bold text-yellow-600">%d</div>
		<div class="text-xs text-yellow-800 font-semibold uppercase tracking-wider mt-1">Warnings</div>
	</div>
	<div class="bg-white p-4 rounded-xl border border-red-100 shadow-sm text-center">
		<div class="text-2xl font-bold text-red-600">%d</div>
		<div class="text-xs text-red-800 font-semibold uppercase tracking-wider mt-1">Broken</div>
	</div>`, len(r.Videos), s.Total, s.Live, s.Warnings, s.Broken)

	var brokenSection string
	if s.Broken == 0 {
		brokenSection = `
	<div class="p-6 bg-green-50 border border-green-200 rounded-2xl text-center">
		<p class="text-green-700 font-semibold">No broken links found!</p>
	</div>`
	} else {
		var critHTML, brokenHTML string
		for _, l := range s.CriticalLinks {
			critHTML += fmt.Sprintf(`
	<div class="bg-white p-3 rounded-lg border border-red-100 text-sm">
		<div class="flex items-center gap-2 mb-1">
			<a href="%s" target="_blank" class="text-red-700 font-mono text-xs break-all hover:underline">%s</a>
		</div>
		<p class="text-gray-600 text-xs">HTTP %d — %s</p>
		<p class="text-gray-500 text-xs mt-1">Video: <a href="https://youtube.com/watch?v=%s" target="_blank" class="text-blue-600 hover:underline">%s</a></p>
	</div>`,
				html.EscapeString(l.URL), html.EscapeString(l.URL),
				l.StatusCode, html.EscapeString(l.Error),
				html.EscapeString(l.VideoID), html.EscapeString(l.VideoTitle),
			)
		}
		for _, l := range s.BrokenLinks {
			brokenHTML += fmt.Sprintf(`
	<div class="bg-white p-3 rounded-lg border border-orange-100 text-sm">
		<div class="flex items-center gap-2 mb-1">
			<a href="%s" target="_blank" class="text-orange-700 font-mono text-xs break-all hover:underline">%s</a>
		</div>
		<p class="text-gray-600 text-xs">HTTP %d — %s</p>
		<p class="text-gray-500 text-xs mt-1">Video: <a href="https://youtube.com/watch?v=%s" target="_blank" class="text-blue-600 hover:underline">%s</a></p>
	</div>`,
				html.EscapeString(l.URL), html.EscapeString(l.URL),
				l.StatusCode, html.EscapeString(l.Error),
				html.EscapeString(l.VideoID), html.EscapeString(l.VideoTitle),
			)
		}
		if critHTML != "" {
			brokenSection += `<div class="p-4 bg-red-50 border border-red-200 rounded-xl"><h3 class="font-bold text-red-800 mb-3">Revenue-Critical Broken Links</h3><div class="space-y-3">` + critHTML + `</div></div>`
		}
		if brokenHTML != "" {
			brokenSection += `<div class="p-4 bg-orange-50 border border-orange-200 rounded-xl"><h3 class="font-bold text-orange-800 mb-3">Broken Links</h3><div class="space-y-3">` + brokenHTML + `</div></div>`
		}
	}

	w.Write([]byte(fmt.Sprintf(`
<div class="space-y-6" id="results">
	<div class="p-6 bg-emerald-50 border border-emerald-200 rounded-2xl shadow-sm text-emerald-800">
		<div class="flex items-center gap-3 mb-4">
			<span class="p-2 bg-emerald-100 text-emerald-600 rounded-xl">
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
			</span>
			<div>
				<h4 class="font-bold text-lg text-emerald-900">Scan Complete</h4>
				<p class="text-xs text-emerald-600 font-semibold tracking-wide uppercase">Channel: %s</p>
			</div>
		</div>
		<div class="pt-4 border-t border-emerald-100/50">
			<div class="bg-white p-4 rounded-xl border border-emerald-100 text-sm space-y-2">
				<div>
					<span class="text-xs font-semibold text-emerald-600 block uppercase tracking-wider">Channel Name</span>
					<strong class="text-emerald-950 font-bold text-base">%s</strong>
				</div>
				<div>
					<span class="text-xs font-semibold text-emerald-600 block uppercase tracking-wider">Channel ID</span>
					<code class="text-emerald-900 font-mono text-xs bg-emerald-50 px-1.5 py-0.5 rounded">%s</code>
				</div>
			</div>
		</div>
	</div>

	<div class="grid grid-cols-2 md:grid-cols-5 gap-4">
		%s
	</div>

	%s

	<div class="p-4 bg-gray-50 border border-gray-200 rounded-xl text-sm text-gray-600">
		<strong>API Quota:</strong> %d used / %d remaining
	</div>
</div>
`,
		html.EscapeString(r.Channel.Name),
		html.EscapeString(r.Channel.Name),
		html.EscapeString(r.Channel.ID),
		statCards,
		brokenSection,
		r.Quota.Used,
		r.Quota.Remaining,
	)))
}

func errorBox(title, message string) string {
	return fmt.Sprintf(`
<div class="p-4 bg-red-50 border border-red-200 rounded-xl text-red-700">
	<div class="flex items-center gap-2 font-bold mb-1 text-red-900">
		<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
		<span>%s</span>
	</div>
	<p class="text-sm">%s</p>
</div>`, html.EscapeString(title), html.EscapeString(message))
}
