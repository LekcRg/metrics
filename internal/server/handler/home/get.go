package home

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/storage"
)

var styles = `<style>
	* {
		margin: 0;
		padding: 0;
	}
	body {
		margin: 8px;
	}

	.title {
		margin-bottom: 16px;
		text-align: center;
	}

	ul, li {
		list-style-type: none;
	}

	.main-list {
		width: 100%;
		display: flex;
	}

	.main-list__item {
		width: 50%;
	}

	.main_list__title {
		margin-bottom: 8px;
		text-align: center;
	}

	.sub-list__item {
		display: flex;
		margin-bottom: 8px;
	}

	.sub-list__name {
		margin-right: 6px;
	}
</style>`

func generateHTMLListItem(name string, value string) string {
	openLiName := `<li class="sub-list__item"><div class="sub-list__name">`
	openDivValue := `:</div><div class="sub-list__value">`
	closeLi := `</div></li>\n`
	itemLen := len(openLiName) + len(openDivValue) + len(closeLi) + len(name) + len(value)
	var res strings.Builder
	res.Grow(itemLen)
	res.WriteString(openLiName)
	res.WriteString(name)
	res.WriteString(openDivValue)
	res.WriteString(value)
	res.WriteString(closeLi)

	return res.String()
}

func generateHTMLList(gaugeList []string, counterList []string) string {
	HTMLList := strings.Join([]string{`
	<li class="main-list__item">
		<h2 class="main_list__title">Gauge</h2>
		<ul class="sub-list">`,
		strings.Join(gaugeList, "\n"),
		`</ul>
	</li>
	<li class="main-list__item">
		<h2 class="main_list__title">Counter</h2>
		<ul class="sub-list">`,
		strings.Join(counterList, "\n"),
		`</ul>
	</li>`}, "")
	return HTMLList
}

func wrapHTML(gaugeList []string, counterList []string) string {
	list := generateHTMLList(gaugeList, counterList)
	HTML := strings.Join([]string{`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Metrics</title>
	</head>
	<body>`,
		styles,
		`<main>
			<h1 class="title">Metrics</h1>
			<ul class="main-list">`,
		list,
		`</ul>
		</main>
	</body>
	</html>`}, "")

	return HTML
}

func generateHTML(list storage.Database) string {
	gaugeList := make([]string, 0, len(list.Gauge))
	counterList := make([]string, 0, len(list.Counter))
	for key, value := range list.Gauge {
		gaugeList = append(gaugeList,
			generateHTMLListItem(key, strconv.FormatFloat(float64(value), 'f', 3, 64)))
	}

	for key, value := range list.Counter {
		counterList = append(counterList,
			generateHTMLListItem(key, strconv.FormatInt(int64(value), 10)))
	}
	return wrapHTML(gaugeList, counterList)
}

func Get(s metric.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := s.GetAllMetrics(r.Context())
		if err != nil {
			http.Error(w, "Internal error 500", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, generateHTML(all))
	}
}
