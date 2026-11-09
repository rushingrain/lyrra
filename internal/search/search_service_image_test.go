package search

import "testing"

func TestNormalizeImageFields(t *testing.T) {
	data := map[string]any{
		"lists": []any{
			map[string]any{
				"lists": []any{
					map[string]any{
						"Image": "http://imge.kugou.com/stdmusic/{size}/cover.jpg",
						"img":   "http://imge.kugou.com/stdmusic/{size}/album.jpg",
						"Name":  "保留原值",
					},
				},
			},
		},
	}

	normalizeImageFields(data)

	items := data["lists"].([]any)
	group := items[0].(map[string]any)
	results := group["lists"].([]any)
	result := results[0].(map[string]any)
	if got, want := result["Image"], "https://imge.kugou.com/stdmusic/400/cover.jpg"; got != want {
		t.Fatalf("Image = %v, want %v", got, want)
	}
	if got, want := result["img"], "https://imge.kugou.com/stdmusic/400/album.jpg"; got != want {
		t.Fatalf("img = %v, want %v", got, want)
	}
	if got, want := result["Name"], "保留原值"; got != want {
		t.Fatalf("Name = %v, want %v", got, want)
	}
}
