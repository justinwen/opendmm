package opendmm

import (
	"sync"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

// MovieMeta contains meta data of movie
type MovieMeta struct {
	Actresses      []string
	ActressTypes   []string
	Categories     []string
	Code           string
	CoverImage     string
	Description    string
	Directors      []string
	Genres         []string
	Label          string
	Maker          string
	MovieLength    string
	Page           string
	ReleaseDate    string
	SampleImages   []string
	Series         string
	Tags           []string
	ThumbnailImage string
	Title          string
}

type searchFunc func(string, *sync.WaitGroup, chan MovieMeta)

// Search for movies based on query and return a channel of MovieMeta
func Search(query string) chan MovieMeta {
	out := make(chan MovieMeta)
	go func(out chan MovieMeta) {
		defer close(out)
		fastOut := searchWithEngines(query, []searchFunc{
			aveSearch,
			caribSearch,
			caribprSearch,
			dmmSearch,
			fc2Search,
			heyzoSearch,
			niceageSearch,
			tkhSearch,
		})
		meta, ok := <-fastOut
		if ok {
			out <- meta
		} else {
			glog.Info("Trying slow engines")
			slowOut := searchWithEngines(query, []searchFunc{
				javSearch,
				mgsSearch,
				opdSearch,
				scuteSearch,
			})
			meta, ok := <-slowOut
			if ok {
				out <- meta
			}
		}
	}(out)
	return out
}

func searchWithEngines(query string, engines []searchFunc) chan MovieMeta {
	wg := new(sync.WaitGroup)
	out := make(chan MovieMeta, 100)
	for _, engine := range engines {
		engine(query, wg, out)
	}
	go func(wg *sync.WaitGroup, out chan MovieMeta) {
		wg.Wait()
		close(out)
	}(wg, out)
	return postprocess(out)
}

// Guess possible movie codes from query string
func Guess(query string) mapset.Set {
	keywords := mapset.NewSet()
	keywords = keywords.Union(aveGuess(query))
	keywords = keywords.Union(caribGuessFull(query))
	keywords = keywords.Union(caribprGuessFull(query))
	keywords = keywords.Union(dmmGuess(query))
	keywords = keywords.Union(heyzoGuessFull(query))
	keywords = keywords.Union(opdGuessFull(query))
	keywords = keywords.Union(tkhGuessFull(query))
	keywords = keywords.Union(scuteGuessFull(query))
	keywords = keywords.Union(fc2GuessFull(query))
	return keywords
}
