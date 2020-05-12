local Millisecond = 1;
local Second = 1000 * Millisecond;
local Minute = 60 * Second;
local Hour = 60 * Minute;
local Day = 24 * Hour;
{
	HTTP: {
		Addr: "127.0.0.1:6000",
		// x509 if empty use h2c
		CertFile: "test.pem",
		KeyFile: "test.key",
	},
	Database:{
		Source:"webpc.db",
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