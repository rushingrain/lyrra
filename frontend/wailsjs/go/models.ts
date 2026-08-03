export namespace main {
	
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

}

