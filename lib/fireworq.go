package mcfireworq

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mackerelio/checkers"
	"github.com/mackerelio/golib/pluginutil"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FireworqCheckApp struct {
	origin string
	tempfile string
}

type FireworqStats struct {
	TotalFailures int64 `json:"total_failures"`
}

func (app *FireworqCheckApp) fetchQueueStats() (map[string]*FireworqStats, error) {
	resp, err := http.Get(app.origin + "/queues/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats map[string]*FireworqStats
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// 各キューについて最新の failure メッセージも取得して返しても良いかもしれない
func (app *FireworqCheckApp) listFailedQueues(stats map[string]*FireworqStats) ([]string, error) {
	var f   *os.File
	var err error

	f, err = os.Open(app.tempfile)
	if os.IsNotExist(err) {
		f, err = os.Create(app.tempfile)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var lastStats map[string]*FireworqStats
	decoder := json.NewDecoder(f)
	decodeErr := decoder.Decode(&lastStats)
	if decodeErr != nil && decodeErr != io.EOF { // 空ファイルだった場合はそのまま処理を継続する
		return nil, err
	}

	var failedQueues []string
	for name, stat := range stats {
		if lastStat, ok := lastStats[name]; ok {
			// すでに監視されているキューの場合は前回から failure が増えているか確認
			if stat.TotalFailures > lastStat.TotalFailures {
				failedQueues = append(failedQueues, name)
			}
		} else {
			// 監視されていないキューなら failure があるか確認
			// 次回からは差分が監視される
			if stat.TotalFailures > 0 {
				failedQueues = append(failedQueues, name)
			}
		}
	}
	return failedQueues, nil
}

func (app *FireworqCheckApp) saveStats(stats map[string]*FireworqStats) error {
	var f   *os.File
	var err error

	f, err = os.OpenFile(app.tempfile, os.O_WRONLY|os.O_TRUNC, 0664)
	if os.IsNotExist(err) {
		f, err = os.Create(app.tempfile)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encodeErr := encoder.Encode(stats)
	if encodeErr != nil {
		return err
	}

	return f.Sync()
}

func Run() *checkers.Checker {
	optName := flag.String("name", "Fireworq", "Name")
	optScheme := flag.String("scheme", "http", "Scheme")
	optHost := flag.String("host", "localhost", "Host")
	optPort := flag.String("port", "8080", "Port")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var tempfile string
	if *optTempfile == "" {
		base := fmt.Sprintf("mackerel-check-fireworq-%s-%s", *optHost, *optPort)
		tempfile = filepath.Join(pluginutil.PluginWorkDir(), base)
	} else {
		tempfile = *optTempfile
	}

	app := FireworqCheckApp {
		origin: fmt.Sprintf("%s://%s:%s", *optScheme, *optHost, *optPort),
		tempfile: tempfile,
	}

	stats, fetchErr := app.fetchQueueStats()
	if fetchErr != nil {
		ckr := checkers.NewChecker(checkers.UNKNOWN, fetchErr.Error())
		ckr.Name = *optName
		return ckr
	}

	failedQueues, listQueueErr := app.listFailedQueues(stats)
	if listQueueErr != nil {
		ckr := checkers.NewChecker(checkers.UNKNOWN, listQueueErr.Error())
		ckr.Name = *optName
		return ckr
	}

	saveErr := app.saveStats(stats)
	if saveErr != nil {
		ckr := checkers.NewChecker(checkers.UNKNOWN, saveErr.Error())
		ckr.Name = *optName
		return ckr
	}

	var checkSt checkers.Status
	var msg string

	if len(failedQueues) == 0 {
		checkSt = checkers.OK
		msg = ""
	} else {
		checkSt = checkers.CRITICAL
		msg = strings.Join(failedQueues, " ")
	}

	ckr := checkers.NewChecker(checkSt, msg)
	ckr.Name = *optName
	return ckr
}
