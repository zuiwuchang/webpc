local Millisecond = 1;
local Second = 1000 * Millisecond;
local Minute = 60 * Second;
local Hour = 60 * Minute;
local Day = 24 * Hour;
{
	HTTP: {
		Addr: "127.0.0.1:9000",
		// x509 if empty use h2c
		// CertFile: "test.pem",
		// KeyFile: "test.key",
	},
	System:{
		// 用戶數據庫
		DB : "webpc.db",
		// 用戶shell
		Shell : ["/bin/bash"],
		// 映射到web的目錄
		Mount : [
			{
				// 網頁上 顯示的 目錄名稱
				Name: "home",
				// 要映射的本地路徑
				Root: "/home/king",
				// 目錄是否可寫
				Write: true,
			},
			{
				Name: "root",
				Root: "/",
				Write: false,
			},
		],
	},
	Cookie: {
		// Filename:"securecookie.json"
		// MaxAge:Day,
	},
	Logger: {
		// zap http
		//HTTP: "localhost:20000",
		// log name
		//Filename:"logs/webpc.log",
		// MB
		MaxSize: 100, 
		// number of files
		MaxBackups: 3,
		// day
		MaxAge: 28,
		// level : debug info warn error dpanic panic fatal
		Level: "debug",
		// output code line and file
		Caller: true,
	},
}