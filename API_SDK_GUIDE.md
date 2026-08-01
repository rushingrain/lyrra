# 酷狗音乐 Go SDK 使用指南

`kugou-music-api` 是酷狗音乐 API 的纯 Go SDK，封装了请求签名、加密、登录态管理等底层逻辑，提供类型安全的 API 调用方式。**框架无关**，可集成到任何 Go 项目中。

---

## 目录

- [1. 安装与导入](#1-安装与导入) — L13
  - [导入方式](#导入方式) — L19
- [2. 快速开始](#2-快速开始) — L33
- [3. Client 初始化与配置](#3-client-初始化与配置) — L68
  - [3.1 基本初始化](#31-基本初始化) — L70
  - [3.2 带选项初始化](#32-带选项初始化) — L77
  - [3.3 全部 Option](#33-全部-option) — L87
  - [3.4 平台模式（Lite vs Standard）](#34-平台模式lite-vs-standard) — L95
  - [3.5 Cookie 管理](#35-cookie-管理) — L109
- [4. 通用 API 调用方式](#4-通用-api-调用方式) — L124
  - [4.1 方式一：类型安全方法](#41-方式一类型安全方法推荐) — L128
  - [4.2 方式二：通用 Call 方法](#42-方式二通用-call-方法) — L142
  - [4.3 Request 结构体字段](#43-request-结构体字段) — L170
- [5. 登录鉴权](#5-登录鉴权) — L186
  - [5.1 登录方式对比](#51-登录方式对比) — L188
  - [5.2 二维码登录](#52-二维码登录三步流程) — L200
  - [5.3 密码登录](#53-密码登录) — L253
  - [5.4 手机验证码登录](#54-手机验证码登录) — L266
  - [5.5 Token 续期登录](#55-token-续期登录) — L281
  - [5.6 会话持久化](#56-会话持久化) — L295
  - [5.7 自定义 Cookie 注入](#57-自定义-cookie-注入) — L344
- [6. API 参考](#6-api-参考) — L366
  - [6.1 Response 结构体](#61-response-结构体) — L368
  - [6.2 API 分类索引](#62-api-分类索引) — L396
    - [搜索](#621-搜索) — L398
    - [歌曲播放地址](#622-歌曲播放地址) — L434
    - [歌词](#623-歌词) — L468
    - [歌单](#624-歌单) — L494
    - [歌手](#625-歌手) — L524
    - [专辑](#626-专辑) — L550
    - [排行榜](#627-排行榜) — L571
    - [推荐与发现](#628-推荐与发现) — L590
    - [用户](#629-用户) — L621
    - [评论](#6210-评论) — L650
    - [视频/MV](#6211-视频mv) — L669
    - [音频/图片](#6212-音频图片) — L679
    - [IP 专区](#6213-ip-专区) — L690
    - [曲谱/歌单广场](#6214-曲谱歌单广场) — L701
    - [长音频（有声书）](#6215-长音频有声书) — L711
    - [酷狗儿歌/Youth](#6216-酷狗儿歌youth) — L722
    - [乐库](#6217-乐库) — L735
    - [其他](#6218-其他) — L743
- [7. 错误处理](#7-错误处理) — L768
  - [7.1 错误类型](#71-错误类型) — L770
  - [7.2 常见业务错误码](#72-常见业务错误码) — L793
  - [7.3 自动续期失败](#73-自动续期失败) — L801
- [8. 进阶用法](#8-进阶用法) — L807
  - [8.1 私人电台（Personal Radio）](#81-私人电台personal-radio) — L809
  - [8.2 歌单保存为电台播放列表](#82-歌单保存为电台播放列表) — L834
  - [8.3 Context 超时控制](#83-context-超时控制) — L844
  - [8.4 并发请求](#84-并发请求) — L854
  - [8.5 完整 API 列表](#85-完整-api-列表) — L871
- [9. 常用 Body 字段提取](#9-常用-body-字段提取) — L882
- [10. 示例代码索引](#10-示例代码索引) — L913
- [11. SDK 文档索引](#11-sdk-文档索引) — L934

---

## 1. 安装与导入

```bash
go get github.com/lfhy/kugou-music-api
```

### 导入方式

```go
// 推荐：直接从根包导入，别名 kg
import kg "github.com/lfhy/kugou-music-api"

// 兼容旧代码：也可导入子包
import "github.com/lfhy/kugou-music-api/sdk"
```

根包通过类型别名暴露了 `sdk` 包的所有公开符号，两者等价。**新项目推荐使用根包**。

---

## 2. 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    kg "github.com/lfhy/kugou-music-api"
)

func main() {
    client, err := kg.New()
    if err != nil {
        log.Fatalf("init client failed: %v", err)
    }

    resp, err := client.Search(context.Background(), kg.SearchRequest{
        Keywords: "周杰伦",
        Page:     1,
        Pagesize: 10,
    })
    if err != nil {
        log.Fatalf("search failed: %v", err)
    }

    fmt.Printf("status: %d\n", resp.Status)
    fmt.Printf("body: %s\n", string(resp.RawBody))
}
```

---

## 3. Client 初始化与配置

### 3.1 基本初始化

```go
// 默认配置：Lite 平台模式，自动续期开启
client, err := kg.New()
```

### 3.2 带选项初始化

```go
client, err := kg.New(
    kg.WithCookie(restoredCookie),   // 注入已保存的 cookie（恢复会话）
    kg.WithLite(false),              // 使用标准平台（非 Lite）
    kg.WithAutoRefresh(true),        // 登录态自动续期（默认已开启）
)
```

### 3.3 全部 Option

| Option | 签名 | 说明 |
|--------|------|------|
| `WithCookie` | `(map[string]string)` | 注入初始 cookie（token / userid 等） |
| `WithAutoRefresh` | `(bool)` | 登录态过期时自动用 token 续期。默认 `true` |
| `WithLite` | `(bool)` | `true`=Lite 平台（默认），`false`=标准平台 |

### 3.4 平台模式（Lite vs Standard）

SDK 默认模拟酷狗 Lite 版客户端。可通过 `WithLite` 或环境变量切换：

```bash
platform=lite    # Lite 平台（默认）
platform=normal  # 标准平台
```

| 模式 | 说明 |
|------|------|
| Lite（默认） | 签名算法较轻量，兼容酷狗概念版 |
| Standard | 与酷狗正式版客户端行为一致，功能更全 |

### 3.5 Cookie 管理

```go
// 获取当前完整 cookie（用于持久化）
cookie := client.Cookie()

// 设置单个 cookie 值
client.SetCookie("token", "xxx")

// 运行时注入额外 cookie
client.SetCookie("dfid", "xxx")
```

---

## 4. 通用 API 调用方式

SDK 提供两种调用方式，覆盖所有 API。

### 4.1 方式一：类型安全方法（推荐）

每个 API 都有对应的强类型 Request / Response 结构和 `Client` 方法，编译期检查参数：

```go
resp, err := client.Search(ctx, kg.SearchRequest{
    Keywords: "周杰伦",
    Page:     1,
    Pagesize: 30,
})
```

完整方法列表见 [第 6 节 API 参考](#6-api-参考)。

### 4.2 方式二：通用 Call 方法

适用于需要动态路由或未生成类型方法的场景：

```go
resp, err := client.Call(ctx, kg.RouteSearch, kg.Request{
    Params: map[string]any{"keywords": "周杰伦", "page": 1, "pagesize": 30},
})
```

可通过 `kg.Endpoints()` 获取全部 API 的路由/方法信息：

```go
for _, ep := range kg.Endpoints() {
    fmt.Printf("%s -> %s %s\n", ep.Identifier, ep.Method, ep.Route)
}
// 输出：
// search -> GET /search
// search_complex -> GET /search/complex
// ...
```

也可按 identifier 反向查路由：

```go
route, ok := kg.RouteByIdentifier("search") // "/search", true
```

### 4.3 Request 结构体字段

```go
type Request struct {
    Params   map[string]any    // URL 查询参数
    Data     map[string]any    // POST body 数据
    Cookie   map[string]string // 本次请求附加 cookie
    Method   string            // HTTP 方法（覆盖默认值）
    URL      string            // 请求路径（覆盖默认值）
    BaseURL  string            // 基础 URL（覆盖默认值）
    Headers  map[string]string // 额外请求头
}
```

---

## 5. 登录鉴权

### 5.1 登录方式对比

| 方式 | 方法 | 适用场景 | 需要登录态 |
|------|------|----------|-----------|
| 二维码登录 | `LoginQrKey` → `LoginQrCreate` → 轮询 `LoginQrCheck` | 桌面客户端、Web 应用 | 否 |
| 密码登录 | `LoginByPassword` | 命令行工具、自动化脚本 | 否 |
| 手机验证码登录 | `CaptchaSent` → `LoginByCellphone` | 移动端 | 否 |
| Token 续期 | `LoginByToken` | 恢复过期会话 | 是（需旧 token） |
| 设备登录 | `LoginDevice` | 设备授权 | 否 |
| 微信登录 | `LoginQrKey` → `LoginWxCreate` → 轮询 `LoginWxCheck` | 微信生态 | 否 |
| 开放平台登录 | `LoginOpenplat` | 第三方接入 | 否 |

### 5.2 二维码登录（三步流程）

```go
// Step 1：获取二维码 key
keyResp, err := client.LoginQrKey(ctx, kg.LoginQrKeyRequest{})
key := keyResp.QRCodeKey()
if key == "" {
    log.Fatal("获取二维码 key 失败:", string(keyResp.RawBody))
}

// Step 2：生成二维码
qrResp, err := client.LoginQrCreate(ctx, kg.LoginQrCreateRequest{
    Key:   key,
    Qrimg: true, // 返回 base64 PNG 图片，false 则只返回 URL
})
qrURL := qrResp.URL()          // 酷狗 H5 扫码页 URL
qrBase64 := qrResp.Base64()    // data:image/png;base64,...
// 将 qrBase64 渲染为二维码图片，或打开 qrURL 供用户扫码

// Step 3：轮询扫码状态（每 2 秒一次）
for {
    time.Sleep(2 * time.Second)
    checkResp, err := client.LoginQrCheck(ctx, kg.LoginQrCheckRequest{Key: key})
    if err != nil {
        continue
    }
    switch checkResp.StatusCode() {
    case 0:
        // 二维码已过期，需重新获取
        return
    case 1:
        // 等待扫码
    case 2:
        // 已扫码，等待用户在手机上确认
    case 4:
        // 登录成功！token/userid 已自动注入 client
        token := checkResp.Token()
        userID := checkResp.UserID()
        // 持久化 cookie
        saveSession(client.Cookie())
        return
    }
}
```

**`LoginQrCheckResponse` 快捷方法**：

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `StatusCode()` | `int` | 0=过期, 1=等待扫码, 2=已扫码待确认, 4=成功 |
| `Token()` | `string` | 登录成功后的 token |
| `UserID()` | `string` | 登录成功后的用户 ID |

### 5.3 密码登录

```go
resp, err := client.LoginByPassword(ctx, kg.PasswordLoginRequest{
    Username: "手机号或用户名",
    Password: "密码",
})
if err != nil {
    log.Fatal("登录失败:", err)
}
// 成功后 token/userid 已自动注入 client
```

### 5.4 手机验证码登录

```go
// Step 1：发送验证码
captchaResp, err := client.CaptchaSent(ctx, kg.CaptchaSentRequest{
    Mobile: "手机号",
})
// Step 2：提交验证码
resp, err := client.LoginByCellphone(ctx, kg.CellphoneLoginRequest{
    Mobile: "手机号",
    Code:   "验证码",
    // UserID: "用户ID", // 有则传，无则留空
})
```

### 5.5 Token 续期登录

用于恢复已保存的会话，或 SDK 自动刷新失败后的手动重试：

```go
resp, err := client.LoginByToken(ctx, kg.TokenLoginRequest{
    Token:  savedToken,
    UserID: savedUserID,
    Cookie: savedCookie, // 可选，传入历史 cookie
})
```

SDK 在任意 API 返回 `error_code=20018`（登录态过期）时，**自动**调用 `LoginByToken` 刷新并重试原请求。可通过 `WithAutoRefresh(false)` 关闭此行为。

### 5.6 会话持久化

需要持久化的关键 cookie 字段：

| Cookie Key | 必须 | 说明 |
|------------|------|------|
| `token` | **是** | 登录凭证 |
| `userid` | **是** | 用户 ID |
| `vip_token` | 否 | VIP 凭证 |
| `vip_type` | 否 | VIP 类型 |

```go
// 保存会话 — 登录成功后调用
func saveSession(cookie map[string]string) {
    essential := map[string]string{
        "token":     cookie["token"],
        "userid":    cookie["userid"],
        "vip_token": cookie["vip_token"],
        "vip_type":  cookie["vip_type"],
    }
    data, _ := json.Marshal(essential)
    os.WriteFile("kugou_session.json", data, 0o600)
}

// 恢复会话 — 初始化 client 时调用
func loadSession() map[string]string {
    data, _ := os.ReadFile("kugou_session.json")
    var cookie map[string]string
    json.Unmarshal(data, &cookie)
    return cookie
}

// 启动时恢复
cookie := loadSession()
client, _ := kg.New(kg.WithCookie(cookie))

// 验证会话是否有效
resp, err := client.UserDetail(ctx, kg.UserDetailRequest{})
if respOK(resp) {
    // 会话有效，进入主逻辑
} else {
    // 尝试 token 续期
    client.LoginByToken(ctx, kg.TokenLoginRequest{
        Token:  cookie["token"],
        UserID: cookie["userid"],
    })
}
```

### 5.7 自定义 Cookie 注入

SDK 在初始化时会自动生成设备标识 cookie：

| Cookie | 来源 |
|--------|------|
| `KUGOU_API_MID` | 基于 `KUGOU_API_GUID` 计算 |
| `KUGOU_API_GUID` | 随机 16 字符 MD5 |
| `KUGOU_API_DEV` | 随机 10 字符 |
| `KUGOU_API_MAC` | 默认 `02:00:00:00:00:00` |
| `KUGOU_API_PLATFORM` | 由平台模式决定 |

可通过环境变量覆盖：

```bash
KUGOU_API_GUID=xxx   # 自定义设备 GUID
KUGOU_API_DEV=xxx    # 自定义设备标识
KUGOU_API_MAC=xxx    # 自定义 MAC
```

---

## 6. API 参考

### 6.1 Response 结构体

所有 API 返回统一的 `*Response` 或具体 Typed Response（本质是 `Response` 的类型别名）：

```go
type Response struct {
    Status  int              // HTTP 状态码
    RawBody []byte           // 原始 JSON 响应体
    Body    map[string]any   // 解析后的 JSON（map）
    Headers map[string]string
    Cookie  []string         // Set-Cookie 响应头
}
```

**业务成功判断**：

```go
func isOK(resp *kg.Response) bool {
    if resp == nil || resp.Body == nil {
        return false
    }
    status  := fmt.Sprintf("%v", resp.Body["status"])
    errCode := fmt.Sprintf("%v", resp.Body["error_code"])
    return (status == "1" || status == "1.0") &&
           (errCode == "" || errCode == "0" || errCode == "<nil>")
}
```

### 6.2 API 分类索引

#### 6.2.1 搜索

| 方法 | 路由 | 说明 |
|------|------|------|
| `Search` | `GET /search` | 综合搜索（歌曲/歌手/专辑） |
| `SearchComplex` | `GET /search/complex` | 高级搜索，更多筛选条件 |
| `SearchMixed` | `GET /search/mixed` | 混合搜索 |
| `SearchSuggest` | `GET /search/suggest` | 搜索建议（即时补全） |
| `SearchHot` | `GET /search/hot` | 热搜词 |
| `SearchDefault` | `POST /search/default` | 默认搜索词 |
| `SearchLyric` | `GET /search/lyric` | 按歌词搜索歌曲 |

```go
// 综合搜索
resp, err := client.Search(ctx, kg.SearchRequest{
    Keywords: "周杰伦",
    Page:     1,
    Pagesize: 30,
})

// 搜索建议
resp, err := client.SearchSuggest(ctx, kg.SearchSuggestRequest{
    Keyword: "周",
})

// 热搜词
resp, err := client.SearchHot(ctx, kg.SearchHotRequest{})

// 歌词搜索
resp, err := client.SearchLyric(ctx, kg.SearchLyricRequest{
    Keywords: "一句歌词",
    Page:     1,
    Pagesize: 30,
})
```

#### 6.2.2 歌曲播放地址

| 方法 | 说明 |
|------|------|
| `GetSongPlayURL` | 获取播放 URL（返回首个可用地址） |
| `ResolveSongPlayURL` | 获取播放 URL 详情（含多地址、来源） |

```go
// 简单用法：获取首个可用播放地址
url, err := client.GetSongPlayURL(ctx, kg.SongPlayURLRequest{
    Hash:         "song_hash",   // 必须，小写
    AlbumID:      123,           // 推荐
    AlbumAudioID: 456,           // 推荐
    Quality:      128,           // 128 | 320 | flac | high
    FreePart:     true,          // 允许试听片段（非 VIP 只有片段）
}, kg.WithSongURLFallback(true)) // 失败时降级到 song_url_new

// 详细用法：获取全部播放地址
result, err := client.ResolveSongPlayURL(ctx, kg.SongPlayURLRequest{...})
// result.Primary  // 首选地址
// result.URLs     // 所有可用地址
// result.Source   // "song_url" 或 "song_url_new"
```

**SongPlayURL 相关 Option**（仅 `GetSongPlayURL` / `ResolveSongPlayURL` 可用）：

| Option | 说明 |
|--------|------|
| `WithSongURLFallback(bool)` | 主接口失败时降级到 `song_url_new`。默认 `true` |
| `WithSongURLAll(bool)` | 返回所有发现的 URL |
| `WithSongURLDFID(string)` | 指定设备 dfid |

> **注意**：`hash` 必须小写。`hash`、`album_id`、`album_audio_id` 通常从搜索/歌单结果中获取。VIP 歌曲需要登录态 + VIP 凭证才能获取高音质。

#### 6.2.3 歌词

| 方法 | 路由 | 说明 |
|------|------|------|
| `Lyric` | `GET /lyric` | 获取歌词 |

```go
resp, err := client.Lyric(ctx, kg.LyricRequest{
    Hash:    "song_hash",
    AlbumID: 123,
})

// 解码后的歌词文本（自动处理 KRC 解密）
text := resp.DecodedContent()

// 转换为标准 LRC 格式（兼容所有音乐播放器）
lrc := resp.ToLrc()
```

`LyricResponse` 提供两个关键方法：

| 方法 | 返回 | 说明 |
|------|------|------|
| `DecodedContent()` | `string` | 解密歌词原文（KRC 格式自动解密为逐字时间轴） |
| `ToLrc()` | `string` | 转换为标准 LRC 格式 |

#### 6.2.4 歌单

| 方法 | 路由 | 说明 |
|------|------|------|
| `PlaylistDetail` | `POST /playlist/detail` | 歌单详情（含歌曲列表） |
| `PlaylistTrackAll` | `GET /playlist/track_all` | 获取歌单所有歌曲 |
| `PlaylistTrackAllNew` | `POST /playlist/track_all_new` | 获取歌单所有歌曲（新版） |
| `PlaylistSimilar` | `POST /playlist/similar` | 相似歌单推荐 |
| `PlaylistTags` | `POST /playlist/tags` | 歌单分类标签 |
| `PlaylistEffect` | `POST /playlist/effect` | 歌单特效（播放量等统计） |
| `PlaylistAdd` | `POST /playlist/add` | 创建歌单 |
| `PlaylistDel` | `POST /playlist/del` | 删除歌单 |
| `PlaylistTracksAdd` | `POST /playlist/tracks/add` | 添加歌曲到歌单 |
| `PlaylistTracksDel` | `POST /playlist/tracks/del` | 从歌单删除歌曲 |

```go
// 歌单详情
resp, err := client.PlaylistDetail(ctx, kg.PlaylistDetailRequest{
    SpecialID: playlistID,
    Page:      1,
    Pagesize:  30,
})

// 创建歌单
resp, err := client.PlaylistAdd(ctx, kg.PlaylistAddRequest{
    Name:   "我的新歌单",
    UserID: userID,
})
```

#### 6.2.5 歌手

| 方法 | 路由 | 说明 |
|------|------|------|
| `ArtistDetail` | `POST /artist/detail` | 歌手详情 |
| `ArtistAlbums` | `POST /artist/albums` | 歌手专辑列表 |
| `ArtistAudios` | `POST /artist/audios` | 歌手歌曲列表 |
| `ArtistVideos` | `GET /artist/videos` | 歌手 MV 列表 |
| `ArtistLists` | `GET /artist/lists` | 歌手分类列表 |
| `ArtistHonour` | `POST /artist/honour` | 歌手荣誉 |
| `ArtistFollow` | `POST /artist/follow` | 关注歌手 |
| `ArtistUnfollow` | `POST /artist/unfollow` | 取消关注 |
| `ArtistFollowNewsongs` | `POST /artist/follow/newsongs` | 关注歌手新歌 |

```go
resp, err := client.ArtistDetail(ctx, kg.ArtistDetailRequest{
    ArtistID: artistID,
})

resp, err := client.ArtistAudios(ctx, kg.ArtistAudiosRequest{
    ArtistID: artistID,
    Page:     1,
    Pagesize: 30,
})
```

#### 6.2.6 专辑

| 方法 | 路由 | 说明 |
|------|------|------|
| `Album` | `POST /album` | 专辑信息 |
| `AlbumDetail` | `POST /album/detail` | 专辑详情 |
| `AlbumSongs` | `POST /album/songs` | 专辑歌曲列表 |
| `AlbumShop` | `GET /album/shop` | 数字专辑商城 |
| `TopAlbum` | `POST /top/album` | 新歌专辑推荐 |

```go
resp, err := client.AlbumDetail(ctx, kg.AlbumDetailRequest{
    AlbumID: albumID,
})

resp, err := client.AlbumSongs(ctx, kg.AlbumSongsRequest{
    AlbumID: albumID,
    Page:    1,
})
```

#### 6.2.7 排行榜

| 方法 | 路由 | 说明 |
|------|------|------|
| `RankList` | `GET /rank/list` | 排行榜分类列表 |
| `RankInfo` | `GET /rank/info` | 排行榜详情 |
| `RankTop` | `GET /rank/top` | 排行榜顶部数据 |
| `RankAudio` | `POST /rank/audio` | 排行榜歌曲列表 |
| `RankVol` | `GET /rank/vol` | 排行榜历史期数 |

```go
resp, err := client.RankList(ctx, kg.RankListRequest{})

resp, err := client.RankAudio(ctx, kg.RankAudioRequest{
    RankID: rankID,
    Page:   1,
})
```

#### 6.2.8 推荐与发现

| 方法 | 路由 | 说明 | 需登录 |
|------|------|------|--------|
| `SceneLists` | `GET /scene/lists` | 场景/首页歌单推荐 | 否 |
| `SceneMusic` | `POST /scene/music` | 场景歌曲推荐 | 否 |
| `SceneModule` | `POST /scene/module` | 首页模块 | 否 |
| `EverydayRecommend` | `POST /everyday/recommend` | 每日推荐 | 是 |
| `EverydayHistory` | `POST /everyday/history` | 每日推荐历史 | 是 |
| `EverydayStyleRecommend` | `POST /everyday/style_recommend` | 风格推荐 | 是 |
| `PersonalFm` | `POST /personal/fm` | 私人 FM 电台 | 是 |
| `FmClass` | `POST /fm/class` | FM 电台分类 | 否 |
| `FmRecommend` | `POST /fm/recommend` | FM 推荐 | 否 |
| `FmSongs` | `POST /fm/songs` | FM 歌曲 | 否 |
| `RecommendSongs` | `POST /recommend/songs` | 推荐歌曲 | 否 |
| `ThemeMusic` | `POST /theme/music` | 主题音乐 | 否 |
| `ThemePlaylist` | `POST /theme/playlist` | 主题歌单 | 否 |

```go
// 每日推荐（需登录）
resp, err := client.EverydayRecommend(ctx, kg.EverydayRecommendRequest{})

// 游客每日推荐（无需登录）
resp, err := client.GetDailyRecommendGuest(ctx, cookie)

// 私人 FM
resp, err := client.PersonalFm(ctx, kg.PersonalFmRequest{})
```

> `GetDailyRecommendGuest` 是手动维护的便捷方法，不是自动生成的 API。通过 `client.GetDailyRecommendGuest(ctx, cookie)` 调用。

#### 6.2.9 用户

| 方法 | 路由 | 说明 | 需登录 |
|------|------|------|--------|
| `UserDetail` | `POST /user/detail` | 用户信息 | 是 |
| `UserPlaylist` | `POST /user/playlist` | 用户歌单 | 是 |
| `UserHistory` | `POST /user/history` | 听歌历史 | 是 |
| `UserListen` | `POST /user/listen` | 收藏/喜欢歌曲 | 是 |
| `UserCloud` | `POST /user/cloud` | 云盘文件列表 | 是 |
| `UserCloudUrl` | `GET /user/cloud/url` | 云盘文件下载地址 | 是 |
| `UserFollow` | `POST /user/follow` | 关注用户 | 是 |
| `UserVipDetail` | `GET /user/vip/detail` | VIP 状态 | 是 |
| `UserVideoCollect` | `POST /user/video/collect` | 用户收藏视频 | 是 |
| `UserVideoLove` | `GET /user/video/love` | 用户喜欢视频 | 是 |
| `FavoriteCount` | `GET /favorite/count` | 收藏数量 | 是 |
| `PlayhistoryUpload` | `POST /playhistory/upload` | 上报播放记录 | 是 |

```go
resp, err := client.UserDetail(ctx, kg.UserDetailRequest{})

resp, err := client.UserPlaylist(ctx, kg.UserPlaylistRequest{
    UserID: userID,
})

resp, err := client.UserHistory(ctx, kg.UserHistoryRequest{
    Page: 1,
})
```

#### 6.2.10 评论

| 方法 | 路由 | 说明 |
|------|------|------|
| `CommentMusic` | `POST /comment/music` | 歌曲评论 |
| `CommentMusicClassify` | `POST /comment/music/classify` | 评论分类 |
| `CommentMusicHotword` | `POST /comment/music/hotword` | 评论热词 |
| `CommentPlaylist` | `POST /comment/playlist` | 歌单评论 |
| `CommentAlbum` | `POST /comment/album` | 专辑评论 |
| `CommentCount` | `GET /comment/count` | 评论数统计 |
| `CommentFloor` | `POST /comment/floor` | 评论楼层/回复 |

```go
resp, err := client.CommentMusic(ctx, kg.CommentMusicRequest{
    Hash: songHash,
    Page: 1,
})
```

#### 6.2.11 视频/MV

| 方法 | 路由 | 说明 |
|------|------|------|
| `VideoDetail` | `POST /video/detail` | 视频/MV 详情 |
| `VideoUrl` | `GET /video/url` | 视频播放地址 |
| `VideoPrivilege` | `POST /video/privilege` | 视频权限信息 |
| `KmrAudioMv` | `POST /kmr/audio/mv` | 歌曲关联 MV |
| `KrmAudio` | `POST /krm/audio` | K 歌伴奏信息 |

#### 6.2.12 音频/图片

| 方法 | 路由 | 说明 |
|------|------|------|
| `Audio` | `POST /audio` | 音频详情 |
| `AudioRelated` | `GET /audio/related` | 相关音频（伴奏等） |
| `AudioAccompanyMatching` | `GET /audio/accompany_matching` | 伴奏匹配 |
| `AudioKtvTotal` | `GET /audio/ktv_total` | K 歌总量 |
| `Images` | `GET /images` | 图片资源（歌手图等） |
| `ImagesAudio` | `GET /images/audio` | 歌曲关联图片 |

#### 6.2.13 IP 专区

| 方法 | 路由 | 说明 |
|------|------|------|
| `Ip` | `POST /ip` | IP 专区列表 |
| `IpDateil` | `POST /ip/dateil` | IP 专区详情 |
| `IpPlaylist` | `POST /ip/playlist` | IP 专区歌单 |
| `IpZone` | `GET /ip/zone` | IP 分区 |
| `IpZoneHome` | `GET /ip/zone/home` | IP 分区首页 |
| `TopIp` | `POST /top/ip` | IP 排行榜 |

#### 6.2.14 曲谱/歌单广场

| 方法 | 路由 | 说明 |
|------|------|------|
| `SheetList` | `GET /sheet/list` | 曲谱列表 |
| `SheetDetail` | `GET /sheet/detail` | 曲谱详情 |
| `SheetHot` | `GET /sheet/hot` | 热门曲谱 |
| `SheetCollection` | `GET /sheet/collection` | 曲谱合集 |
| `SheetCollectionDetail` | `GET /sheet/collection/detail` | 曲谱合集详情 |

#### 6.2.15 长音频（有声书）

| 方法 | 路由 | 说明 |
|------|------|------|
| `LongaudioAlbumDetail` | `POST /longaudio/album/detail` | 有声书专辑详情 |
| `LongaudioAlbumAudios` | `POST /longaudio/album/audios` | 有声书音频列表 |
| `LongaudioDailyRecommend` | `POST /longaudio/daily_recommend` | 有声书每日推荐 |
| `LongaudioRankRecommend` | `GET /longaudio/rank_recommend` | 有声书排行推荐 |
| `LongaudioVipRecommend` | `POST /longaudio/vip_recommend` | 有声书 VIP 推荐 |
| `LongaudioWeekRecommend` | `POST /longaudio/week_recommend` | 有声书周推荐 |

#### 6.2.16 酷狗儿歌/Youth

| 方法 | 路由 | 说明 |
|------|------|------|
| `YouthChannelAll` | `GET /youth/channel/all` | 儿歌频道列表 |
| `YouthChannelDetail` | `POST /youth/channel/detail` | 儿歌频道详情 |
| `YouthChannelSong` | `GET /youth/channel/song` | 儿歌频道歌曲 |
| `YouthChannelSub` | `GET /youth/channel/sub` | 儿歌订阅频道 |
| `YouthDynamic` | `GET /youth/dynamic` | 儿歌动态 |
| `YouthListenSong` | `POST /youth/listen/song` | 儿歌播放 |
| `YouthVip` | `POST /youth/vip` | 儿歌 VIP |
| `TopCardYouth` | `POST /top/card/youth` | 儿歌首页卡片 |

#### 6.2.17 乐库

| 方法 | 路由 | 说明 |
|------|------|------|
| `Yueku` | `GET /yueku` | 乐库首页 |
| `YuekuBanner` | `POST /yueku/banner` | 乐库 Banner |
| `YuekuFm` | `GET /yueku/fm` | 乐库 FM |

#### 6.2.18 其他

| 方法 | 路由 | 说明 |
|------|------|------|
| `ServerNow` | `POST /server/now` | 服务器时间 |
| `RegisterDev` | `POST /register/dev` | 注册设备 |
| `Brush` | `POST /brush` | 刷量/曝光上报 |
| `TopCard` | `POST /top/card` | 首页卡片 |
| `TopSong` | `POST /top/song` | 歌曲排行 |
| `TopPlaylist` | `POST /top/playlist` | 歌单排行 |
| `SongRanking` | `GET /song/ranking` | 歌曲排名 |
| `SongRankingFilter` | `GET /song/ranking/filter` | 歌曲排名筛选 |
| `SongClimax` | `GET /song/climax` | 歌曲高潮片段 |
| `LastestSongsListen` | `POST /lastest/songs/listen` | 最新歌曲试听 |
| `PcDiantai` | `POST /pc/diantai` | PC 电台 |
| `PrivilegeLite` | `POST /privilege/lite` | 权限信息（Lite） |
| `SceneAudioList` | `POST /scene/audio_list` | 场景音频列表 |
| `SceneCollectionList` | `POST /scene/collection_list` | 场景合集列表 |
| `SceneVideoList` | `POST /scene/video_list` | 场景视频列表 |
| `SceneModuleInfo` | `GET /scene/module_info` | 场景模块信息 |
| `AiRecommend` | `POST /ai/recommend` | AI 推荐 |
| `SingerList` | `GET /singer/list` | 歌手分类列表 |

---

## 7. 错误处理

### 7.1 错误类型

```go
resp, err := client.Search(ctx, req)
if err != nil {
    // 类型 1：网络错误 / SDK 内部错误（签名失败、token 刷新失败等）
    log.Printf("请求失败: %v", err)
    return
}
// 类型 2：HTTP 非 2xx（resp.Status != 200）
if resp.Status != 200 {
    log.Printf("HTTP 错误: %d", resp.Status)
    return
}
// 类型 3：业务错误（酷狗 API 返回 error_code != 0）
if !isBizOK(resp) {
    errCode := resp.Body["error_code"]
    errMsg  := resp.Body["msg"]
    log.Printf("业务错误: code=%v msg=%v", errCode, errMsg)
    return
}
```

### 7.2 常见业务错误码

| error_code | 说明 | 处理 |
|------------|------|------|
| `0` | 成功 | - |
| `20018` | 登录态过期 | SDK 自动 token 刷新（需 `WithAutoRefresh(true)`） |
| 其他 | 参数错误/无权限/服务器错误等 | 查看 `msg` / `error` 字段 |

### 7.3 自动续期失败

当 `WithAutoRefresh(true)`（默认）但 token 刷新失败时，SDK 返回包含 `errLoginExpired` 的错误。此时需要提示用户重新登录。

---

## 8. 进阶用法

### 8.1 私人电台（Personal Radio）

SDK 提供内置的三种电台模式 + 通用电台请求，封装了 K 歌、新歌、小众等场景：

```go
// 方式 1：内置模式
tracks, err := client.PersonalRadio(ctx, kg.PersonalRadioHeart) // K 歌电台
tracks, err := client.PersonalRadio(ctx, kg.PersonalRadioNew)   // 新歌电台
tracks, err := client.PersonalRadio(ctx, kg.PersonalRadioNiche) // 小众电台

// 方式 2：自定义参数
resp, err := client.PersonalRadioRequest(ctx, kg.PersonalRadioRequest{
    Mode: "heart",
})

// 返回 []RadioTrack，包含 hash / album_id 等信息，可直接传给 GetSongPlayURL
for _, track := range tracks {
    url, _ := client.GetSongPlayURL(ctx, kg.SongPlayURLRequest{
        Hash:         track.Hash,
        AlbumID:      track.AlbumID,
        AlbumAudioID: track.AlbumAudioID,
    })
}
```

### 8.2 歌单保存为电台播放列表

```go
playlist, err := client.PlaylistDetail(ctx, kg.PlaylistDetailRequest{
    SpecialID: playlistID,
    Pagesize:  100,
})
// 从 resp.Body 提取歌曲列表，逐首获取播放地址后保存
```

### 8.3 Context 超时控制

所有 API 方法接受 `context.Context`，可设置超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := client.Search(ctx, kg.SearchRequest{...})
```

### 8.4 并发请求

`Client` 实例是并发安全的，可在多个 goroutine 中同时使用：

```go
var wg sync.WaitGroup
for _, hash := range songHashes {
    wg.Add(1)
    go func(h string) {
        defer wg.Done()
        url, _ := client.GetSongPlayURL(ctx, kg.SongPlayURLRequest{Hash: h})
        // ...
    }(hash)
}
wg.Wait()
```

### 8.5 完整 API 列表

```go
// 打印所有可用 API 及其路由
for _, ep := range client.Endpoints() {
    fmt.Printf("%-30s %-6s %s\n", ep.Identifier, ep.Method, ep.Route)
}
```

---

## 9. 常用 Body 字段提取

酷狗 API 响应 Body 是 `map[string]any`，字段命名不统一。常用提取模式：

```go
// 提取 data 对象
data, _ := resp.Body["data"].(map[string]any)

// 提取列表 — 常见 key: info / list / data.infos
list, _ := data["info"].([]any)
if list == nil {
    list, _ = data["list"].([]any)
}

// 提取歌曲信息
for _, item := range list {
    song := item.(map[string]any)
    name := fmt.Sprintf("%v", song["songname"])
    hash := fmt.Sprintf("%v", song["hash"])
    albumID := toInt(song["album_id"])
    albumAudioID := toInt(song["album_audio_id"])
}

// 提取播放地址 — 常见 key: url / play_url / backup_url
url := findFirstURL(resp.Body)
```

参考 `examples/daily_recommend/main.go` 中的 `extractSongs` / `extractSongURL` 等辅助函数。

---

## 10. 示例代码索引

`examples/` 目录包含完整可运行示例：

| 示例 | 目录 | 覆盖内容 |
|------|------|----------|
| 最小搜索 | `simple/` | 客户端初始化、搜索调用 |
| 二维码登录 | `login_qrcode/` | 完整三步登录、终端二维码渲染、会话保存 |
| 会话管理 | `login_session/` | 会话持久化+恢复、密码登录、token 续期 |
| 每日推荐 | `daily_recommend/` | 推荐获取、播放地址解析、Body 字段提取 |
| 手机验证码登录 | `login_cellphone/` | 验证码发送 + 登录 |
| 创建歌单 | `create_playlist/` | 歌单创建 |
| 添加歌曲到歌单 | `add_playlist_tracks/` | 歌单编辑 |
| 私人电台 | `personal_radio/` | 电台获取 + 歌单保存 |
| 歌词转 LRC | `lyric_to_lrc/` | 歌词解码与格式转换 |
| Token 刷新 | `token_refresh/` | 手动 token 续期 |
| 用户信息刷新 | `user_info_refresh/` | 用户信息同步 |
| 自动每日签到 | `auto_daily_checkin/` | 签到自动化 |

---

## 11. SDK 文档索引

| 文档 | 路径 | 内容 |
|------|------|------|
| API 清单（英文） | `sdk/API_CATALOG.md` | 全量 API 路由/方法/模型 |
| API 清单（中文） | `sdk/API_CATALOG_ZH.md` | 中文说明（从原项目注释提取） |
| 兼容性审计 | `sdk/API_COMPAT_AUDIT.md` | 逐接口兼容风险 |
| 修复状态 | `sdk/API_FIX_STATUS.md` | 逐接口校对进度 |
| 签名检查 | `sdk/API_SIGN_CHECK.md` | 逐接口签名/加密检查 |
| 完整 README | `README.md` | 项目总览 + 调用函数清单 |
