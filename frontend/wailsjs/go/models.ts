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

export namespace banner {
	
	export class HomeBanner {
	    id: string;
	    imageUrl: string;
	    title: string;
	    linkUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new HomeBanner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.imageUrl = source["imageUrl"];
	        this.title = source["title"];
	        this.linkUrl = source["linkUrl"];
	    }
	}

}

export namespace playlist {
	
	export class HomePlaylist {
	    id: string;
	    name: string;
	    coverUrl: string;
	    description: string;
	    playCount: number;
	
	    static createFrom(source: any = {}) {
	        return new HomePlaylist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.coverUrl = source["coverUrl"];
	        this.description = source["description"];
	        this.playCount = source["playCount"];
	    }
	}

}

export namespace song {
	
	export class HomeSong {
	    id: string;
	    title: string;
	    artist: string;
	    album: string;
	    coverUrl: string;
	    duration: number;
	    hash: string;
	    quality: string;
	    isOriginal: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HomeSong(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.coverUrl = source["coverUrl"];
	        this.duration = source["duration"];
	        this.hash = source["hash"];
	        this.quality = source["quality"];
	        this.isOriginal = source["isOriginal"];
	    }
	}

}

