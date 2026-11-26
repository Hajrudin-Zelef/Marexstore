package render

import (
	"os"
	"strconv"
	"sync"
)

// imgV : versionne l'URL d'une image produit avec sa date de modification
// (?v=<mtime>). Quand un lot d'images remplace un fichier, l'URL change et
// les navigateurs récupèrent la nouvelle image immédiatement, malgré le
// Cache-Control max-age=86400 du handler /images/. La table est lue une
// seule fois (le conteneur redémarre à chaque déploiement). Fichier absent
// (tests, aperçu) : URL inchangée.
var imgVerOnce sync.Once
var imgVer map[string]int64

func imgV(p string) string {
	imgVerOnce.Do(func() {
		imgVer = map[string]int64{}
		entries, err := os.ReadDir("images")
		if err != nil {
			return
		}
		for _, e := range entries {
			if info, err := e.Info(); err == nil && !e.IsDir() {
				imgVer["images/"+e.Name()] = info.ModTime().Unix()
			}
		}
	})
	if v, ok := imgVer[p]; ok {
		return p + "?v=" + strconv.FormatInt(v, 10)
	}
	return p
}
