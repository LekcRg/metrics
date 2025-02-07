package home

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	return fmt.Sprintf(`<li class="sub-list__item">
	<div class="sub-list__name">%s:</div>
	<div class="sub-list__value">%s</div>
</li>`, name, value)
}

func generateHTMLList(gaugeList []string, counterList []string) string {
	HTMLList := fmt.Sprintf(`<li class="main-list__item">
	<h2 class="main_list__title">Gauge</h2>
	<ul class="sub-list">
		%s
	</ul>
</li>
<li class="main-list__item">
	<h2 class="main_list__title">Counter</h2>
	<ul class="sub-list">
		%s
	</ul>
</li>`, strings.Join(gaugeList, "\n"), strings.Join(counterList, "\n"))
	return HTMLList
}

func generateHTML(gaugeList []string, counterList []string) string {
	list := generateHTMLList(gaugeList, counterList)
	HTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Metrics</title>
</head>
<body>
	%s
	<main>
		<h1 class="title">Metrics</h1>
		<ul class="main-list">
			%s
		</ul>
	</main>
</body>
</html>`, styles, list)

	return HTML
}

type database interface {
	GetAll() (storage.Database, error)
}

func Get(db database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, err := db.GetAll()
		if err != nil {
			http.Error(w, "Internal error 500", http.StatusInternalServerError)
		}
		gaugeList := []string{}
		counterList := []string{}
		for key, value := range all.Gauge {
			gaugeList = append(gaugeList,
				generateHTMLListItem(key, strconv.FormatFloat(float64(value), 'f', -1, 64)))
		}

		for key, value := range all.Counter {
			counterList = append(counterList,
				generateHTMLListItem(key, fmt.Sprintf("%d", value)))
		}

		result := generateHTML(gaugeList, counterList)
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, result)
	}
}
