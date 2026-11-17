package handlers

import (
	"reflect"
	"testing"
)

// mergeGallery : extras conservés (moins suppressions) + nouveaux,
// plafond 4 extras (5 photos avec la principale).
func TestMergeGaleriePlafondQuatreExtras(t *testing.T) {
	kept := []string{"images/p-2.jpg", "images/p-3.jpg"}
	del := map[string]bool{"images/p-2.jpg": true}
	up := []string{"images/p-4.jpg", "images/p-5.jpg", "images/p-6.jpg", "images/p-7.jpg"}
	want := []string{"images/p-3.jpg", "images/p-4.jpg", "images/p-5.jpg", "images/p-6.jpg"}
	if got := mergeGallery(kept, del, up); !reflect.DeepEqual(got, want) {
		t.Fatalf("merge : %v, attendu %v", got, want)
	}
}

func TestMergeGalerieSansDoublonNiVide(t *testing.T) {
	got := mergeGallery([]string{"images/a.jpg"}, nil, []string{"", "images/a.jpg", "images/b.jpg"})
	want := []string{"images/a.jpg", "images/b.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge : %v, attendu %v", got, want)
	}
}
