package main

const (
	diagSimple = "Diagnostics - Simple"
	diagFull   = "Diagnostics - Full"
	diagCustom = "Diagnostics - Custom"

	running    = "running"
	loopback   = "loopback"
	resolvConf = "/etc/resolv.conf"

	cacheRepo       = "https://raw.githubusercontent.com/uklans/cache-domains/master/"
	heartbeatSuffix = "/lancache-heartbeat"
	httpPrefix      = "http://"
	lancacheHeader  = "X-Lancache-Processed-By"
	testHostname    = "lancache.steamcontent.com"

	testPrefix     = "lancachetest."
	wildcardPrefix = "*."

	portHTTP = ":80"
	portDNS  = ":53"
)

var (
	CDNs = []CDN{ArenaNet, Blizzard, BattleStateGames, CallOfDuty, CityOfHeroes, DaybreakGames, EpicGames, Frontier, Neverwinter,
		NexusMods, Nintendo, Origin, PathOfExile, RenegadeX, RiotGames, RockstarGames, Sony, SquareEnix, Steam, Test,
		TheElderScrollsOnline, UPlay, Warframe, Wargaming, WindowsUpdates, XboxLive}

	ArenaNet = CDN{
		Name: "ArenaNet",
		File: "arenanet.txt",
	}
	Blizzard = CDN{
		Name: "Blizzard",
		File: "blizzard.txt",
	}
	BattleStateGames = CDN{
		Name: "Battle State Games",
		File: "bsg.txt",
	}
	CallOfDuty = CDN{
		Name: "Call of Duty",
		File: "cod.txt",
	}
	CityOfHeroes = CDN{
		Name: "City of Heroes",
		File: "cityofheroes.txt",
	}
	DaybreakGames = CDN{
		Name: "Daybreak Games",
		File: "daybreak.txt",
	}
	EpicGames = CDN{
		Name: "Epic Games",
		File: "epicgames.txt",
	}
	Frontier = CDN{
		Name: "Frontier",
		File: "frontier.txt",
	}
	Neverwinter = CDN{
		Name: "Neverwinter",
		File: "neverwinter.txt",
	}
	NexusMods = CDN{
		Name: "Nexus Mods",
		File: "nexusmods.txt",
	}
	Nintendo = CDN{
		Name: "Nintendo",
		File: "nintendo.txt",
	}
	Origin = CDN{
		Name: "Origin",
		File: "origin.txt",
	}
	PathOfExile = CDN{
		Name: "Path of Exile",
		File: "pathofexile.txt",
	}
	RenegadeX = CDN{
		Name: "RenegadeX",
		File: "renegadex.txt",
	}
	RiotGames = CDN{
		Name: "Riot Games",
		File: "riot.txt",
	}
	RockstarGames = CDN{
		Name: "Rockstar Games",
		File: "rockstar.txt",
	}
	Sony = CDN{
		Name: "Sony",
		File: "sony.txt",
	}
	SquareEnix = CDN{
		Name: "SQUARE ENIX",
		File: "square.txt",
	}
	Steam = CDN{
		Name: "Steam",
		File: "steam.txt",
	}
	Test = CDN{
		Name: "Test",
		File: "test.txt",
	}
	TheElderScrollsOnline = CDN{
		Name: "The Elder Scrolls Online",
		File: "teso.txt",
	}
	UPlay = CDN{
		Name: "UPlay",
		File: "uplay.txt",
	}
	Warframe = CDN{
		Name: "Warframe",
		File: "warframe.txt",
	}
	Wargaming = CDN{
		Name: "WARGAMING",
		File: "wargaming.net.txt",
	}
	WindowsUpdates = CDN{
		Name: "Windows Updates",
		File: "windowsupdates.txt",
	}
	XboxLive = CDN{
		Name: "Xbox Live",
		File: "xboxlive.txt",
	}

	systemResolver = []string{"system"}
)
