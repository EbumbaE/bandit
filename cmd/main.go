package cmd

import (
	"sync"

	bandit_indexer "github.com/EbumbaE/bandit/services/bandit-indexer/cmd/run"
	rule_admin "github.com/EbumbaE/bandit/services/rule-admin/cmd/run"
	rule_diller "github.com/EbumbaE/bandit/services/rule-diller/cmd/run"
)

func main() {
	runables := []func(){
		rule_admin.Run,
		rule_diller.Run,
		bandit_indexer.Run,
	}

	wg := sync.WaitGroup{}
	wg.Add(len(runables))

	for _, run := range runables {
		go func() {
			run()
			wg.Done()
		}()
	}

	wg.Wait()
}
