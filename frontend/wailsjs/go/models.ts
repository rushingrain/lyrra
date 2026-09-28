export namespace auth {
	
	export class QrLoginResponse {
	    qrKey: string;
	    qrUrl: string;
	    expiresAt: number;
	
	    static createFrom(source: any = {}) {
	        return new QrLoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.qrKey = source["qrKey"];
	        this.qrUrl = source["qrUrl"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class UserInfo {
	    nickname: string;
	    avatar: string;
	    isVip: boolean;
	    friends: number;
	    follows: number;
	    fans: number;
	
	    static createFrom(source: any = {}) {
	        return new UserInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nickname = source["nickname"];
	        this.avatar = source["avatar"];
	        this.isVip = source["isVip"];
	        this.friends = source["friends"];
	        this.follows = source["follows"];
	        this.fans = source["fans"];
	    }
	}
	export class UserPlaylist {
	    id: string;
	    name: string;
	    songCount: number;
	    coverUrl: string;
	    isDefault: number;
	    type: number;
	
	    static createFrom(source: any = {}) {
	        return new UserPlaylist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.songCount = source["songCount"];
	        this.coverUrl = source["coverUrl"];
	        this.isDefault = source["isDefault"];
	        this.type = source["type"];
	    }
	}

}

export namespace dto {
	
	export class Album {
	    id: number;
	    name: string;
	    artist: string;
	    coverUrl: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new Album(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.artist = source["artist"];
	        this.coverUrl = source["coverUrl"];
	        this.date = source["date"];
	    }
	}
	export class Artist {
	    id: number;
	    name: string;
	    avatarUrl: string;
	    songCount: number;
	    genre: string;
	    fans: number;
	
	    static createFrom(source: any = {}) {
	        return new Artist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.avatarUrl = source["avatarUrl"];
	        this.songCount = source["songCount"];
	        this.genre = source["genre"];
	        this.fans = source["fans"];
	    }
	}
	export class Banner {
	    id: string;
	    imageUrl: string;
	    title: string;
	    linkUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Banner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.imageUrl = source["imageUrl"];
	        this.title = source["title"];
	        this.linkUrl = source["linkUrl"];
	    }
	}
	export class MV {
	    id: number;
	    name: string;
	    artist: string;
	    coverUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new MV(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.artist = source["artist"];
	        this.coverUrl = source["coverUrl"];
	    }
	}
	export class Playlist {
	    id: number;
	    name: string;
	    coverUrl: string;
	    description: string;
	    creator: string;
	    playCount: number;
	    songCount?: number;
	
	    static createFrom(source: any = {}) {
	        return new Playlist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.coverUrl = source["coverUrl"];
	        this.description = source["description"];
	        this.creator = source["creator"];
	        this.playCount = source["playCount"];
	        this.songCount = source["songCount"];
	    }
	}
	export class Song {
	    id: number;
	    name: string;
	    artist: string;
	    album?: string;
	    coverUrl: string;
	    duration?: number;
	    hash?: string;
	    quality?: string;
	    isOriginal: boolean;
	    tags?: string[];
	    matchedLyric?: string;
	
	    static createFrom(source: any = {}) {
	        return new Song(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.coverUrl = source["coverUrl"];
	        this.duration = source["duration"];
	        this.hash = source["hash"];
	        this.quality = source["quality"];
	        this.isOriginal = source["isOriginal"];
	        this.tags = source["tags"];
	        this.matchedLyric = source["matchedLyric"];
	    }
	}

}

export namespace play {
	
	export class PlayURL {
	    url: string;
	    quality: string;
	    qualityParam: string;
	    fileHash?: string;
	    degraded: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlayURL(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.quality = source["quality"];
	        this.qualityParam = source["qualityParam"];
	        this.fileHash = source["fileHash"];
	        this.degraded = source["degraded"];
	    }
	}

}

export namespace search {
	
	export class Result {
	    songs: dto.Song[];
	    playlists: dto.Playlist[];
	    albums: dto.Album[];
	    artists: dto.Artist[];
	    mvs: dto.MV[];
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.songs = this.convertValues(source["songs"], dto.Song);
	        this.playlists = this.convertValues(source["playlists"], dto.Playlist);
	        this.albums = this.convertValues(source["albums"], dto.Album);
	        this.artists = this.convertValues(source["artists"], dto.Artist);
	        this.mvs = this.convertValues(source["mvs"], dto.MV);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

