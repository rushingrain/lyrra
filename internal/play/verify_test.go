package play

import "testing"

func TestFileHashFromPlayURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "标准音质 URL",
			url:  "http://fs.youthandroid2.kugou.com/202609271657/661b3a42/v3/1DA15BA0B42B8DF19C27FC99A04B7829/yp/full/ap3116_us1_pi411_mx0_qu128_pa1_ct1_s1.mp3",
			want: "1da15ba0b42b8df19c27fc99a04b7829",
		},
		{
			name: "无损 URL",
			url:  "http://fs.x.kugou.com/202609271657/sig/v3/08775206d7cc07a363967d40a00e493c/yp/full/quflac_pa1_s1.flac",
			want: "08775206d7cc07a363967d40a00e493c",
		},
		{name: "封面图", url: "http://imge.kugou.com/stdmusic/400/20250125/cover.jpg", want: ""},
		{name: "空地址", url: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileHashFromPlayURL(tt.url); got != tt.want {
				t.Fatalf("fileHashFromPlayURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsImageURL(t *testing.T) {
	if !isImageURL("http://imge.kugou.com/stdmusic/{size}/20250125/20250125121653941441.jpg") {
		t.Fatal("jpg 应判定为图片")
	}
	if isImageURL("http://fs.x.kugou.com/v3/abc/yp/full/qu128_a.mp3") {
		t.Fatal("mp3 不应判定为图片")
	}
}

func TestBuildPlayURL(t *testing.T) {
	const (
		expectedHash = "08775206D7CC07A363967D40A00E493C"
		matchedURL   = "http://fs.x.kugou.com/202609271657/sig/v3/08775206d7cc07a363967d40a00e493c/yp/full/quflac_a.flac"
		// 典型降级：歌曲没有 Hi-Res 时，high 请求返回标准文件（标记仍是 quhigh）
		degradedURL = "http://fs.x.kugou.com/202609271657/sig/v3/1da15ba0b42b8df19c27fc99a04b7829/yp/full/quhigh_a.mp3"
		coverURL    = "http://imge.kugou.com/stdmusic/400/20250125/cover.jpg"
	)

	t.Run("命中目标档位", func(t *testing.T) {
		result, err := buildSongPlayURL("sq", "flac", expectedHash, matchedURL)
		if err != nil {
			t.Fatalf("buildSongPlayURL() error = %v", err)
		}
		if result.Degraded || result.FileHash != "08775206d7cc07a363967d40a00e493c" {
			t.Fatalf("unexpected result: %+v", result)
		}
		if result.Quality != "sq" || result.QualityParam != "flac" || result.URL != matchedURL {
			t.Fatalf("unexpected result: %+v", result)
		}
	})

	t.Run("被上游降级", func(t *testing.T) {
		result, err := buildSongPlayURL("res", "high", expectedHash, degradedURL)
		if err != nil {
			t.Fatalf("buildSongPlayURL() error = %v", err)
		}
		if !result.Degraded {
			t.Fatalf("应标记为降级: %+v", result)
		}
		if result.FileHash != "1da15ba0b42b8df19c27fc99a04b7829" {
			t.Fatalf("fileHash = %q", result.FileHash)
		}
	})

	t.Run("无地址", func(t *testing.T) {
		if _, err := buildSongPlayURL("sq", "flac", expectedHash, ""); err == nil {
			t.Fatal("空地址应返回错误")
		}
	})

	t.Run("SDK 把封面图当地址", func(t *testing.T) {
		if _, err := buildSongPlayURL("res", "high", expectedHash, coverURL); err == nil {
			t.Fatal("图片地址应返回错误")
		}
	})
}

func TestBuildPlayURLUsesURLQualityWhenSongHashIsStandard(t *testing.T) {
	url := "http://fs.x.kugou.com/202609271657/sig/v3/08775206d7cc07a363967d40a00e493c/yp/full/quflac_a.flac"
	result, err := buildSongPlayURL(LevelSQ, "flac", "standard-song-hash", url)
	if err != nil {
		t.Fatalf("buildSongPlayURL() error = %v", err)
	}
	if result.Degraded {
		t.Fatalf("matching URL quality should not be marked degraded: %+v", result)
	}
}
