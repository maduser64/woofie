-- Client program for the Woofie virtual dog.

-- This program is intended to be used on an ESP8266 board that is hooked up
-- to a motion sensor.  The idea is that, when motion is detected, the wifi
-- is fired up and a UDP packet is sent to the broadcast IP on a given port
-- to trigger the server to bark the virtual dog.

-- Due to interference problems with the wifi potentially causing false
-- positives, we leave the wifi off under normal operation.  When motion is
-- detected, we turn wifi back on long enough to get an IP and transmit the
-- packet.  Then we leave it on for up to CONFIG_WIFITIME milliseconds in case
-- there's more data to send.

-- (C) 2017 BJ Black <bj@wjblack.com>

-- Released under the WTFPL.  See COPYING in the woofie source tree for details



-- Config

-- The GPIO button that the config button is hooked up to.
CONFIG_BTN = 3

-- The GPIO that the motion sensor is hooked up to.
CONFIG_MOTION = 7

-- The preshared key for the packet generation.
CONFIG_PASSWORD = "bow wow"

-- The UDP port number to broadcast to
CONFIG_PORT = 40080

-- How many milliseconds to leave smartconfig on after button press if it
-- doesn't get a successful config.
CONFIG_SETUPTIME = 30000

-- How many milliseconds to leave the Wifi on after first packet sent (to keep
-- from getting too many up/downs on the interface).
CONFIG_WIFITIME = 5000


-- Precompute the on and off packets (we don't use the off packets yet)
onpacket = crypto.hash("md5", CONFIG_PASSWORD .. ":on")
offpacket = crypto.hash("md5", CONFIG_PASSWORD .. ":off")


-- Set the RTC to a nonzero value for now
rtctime.set(1000000000,0)

-- timer IDs we'll be using (OO timers are a titch wonky :-/ )
TIMER_SETUP = 1
TIMER_TEARDOWN = 2

-- Tear down the Wifi after the packets are sent
function teardown()
	print("Woofs sent.  Tearing down WiFi...")
	reset()
	wifi.setmode(wifi.NULLMODE, true)
	tmr.register(TIMER_TEARDOWN, CONFIG_WIFITIME, tmr.ALARM_SINGLE,
		startup)
	tmr.start(TIMER_TEARDOWN)
end

-- Callback for sending the motion sensor packet once we get an IP.
function sendmotion(previous_state)
	ip = wifi.sta.getip()
	bip = wifi.sta.getbroadcast()
	print("Got IP " .. ip .. ", so sending woof request to " .. bip .. "...")
	-- Figure out what the broadcast address is to send the packet.
	sock = net.createUDPSocket()
	sock:send(CONFIG_PORT, bip, onpacket)
	sock:send(CONFIG_PORT, bip, onpacket)
	if tmr.state(TIMER_TEARDOWN) == nil then
		tmr.register(TIMER_TEARDOWN, CONFIG_WIFITIME, tmr.ALARM_SINGLE,
			teardown)
		tmr.start(TIMER_TEARDOWN)
	end
end

-- Callback for the motion sensor.
function firemotion()
	print("Detected motion...")
	reset()
	wifi.sta.eventMonReg(wifi.STA_GOTIP, sendmotion)
	wifi.sta.eventMonStart()
	tmr.register(TIMER_SETUP, CONFIG_SETUPTIME, tmr.ALARM_SINGLE,
		startup)
	tmr.start(TIMER_SETUP)
	wifi.setmode(wifi.STATION, false)
	wifi.sta.connect()
	print("Waiting for IP...")
end

-- Set up an edge trigger to respond to the pushbutton (for smartconfig).
-- We set the GPIO trigger back up at the end for the next button press.
function setconfig(ssid, password)
	print("Got smartconfig " .. ssid .. ", " .. password .. "  Configuring...")
	reset()
	wifi.stopsmart()
	config = {}
	config.ssid = ssid
	config.pwd = password
	config.auto = false
	config.save = true
	wifi.sta.config(config)
	wifi.sta.eventMonReg(wifi.STA_GOTIP, sendmotion)
	wifi.sta.eventMonStart()
	tmr.register(TIMER_SETUP, CONFIG_SETUPTIME, tmr.ALARM_SINGLE,
		setuptimeout)
	tmr.start(TIMER_SETUP)
end

-- Callback when setup has been running without success for too long.
function setuptimeout()
	print("Setup timed out.")
	wifi.stopsmart()
	startup()
end

-- Callback when we get a config pushbutton down.
function fireconfig()
	print("Got buttonpress...")
	reset()
	wifi.setmode(wifi.STATION)
	tmr.register(TIMER_SETUP, CONFIG_SETUPTIME, tmr.ALARM_SINGLE,
		setuptimeout)
	tmr.start(TIMER_SETUP)
	wifi.startsmart(0, setconfig)
end

-- Configured state.  Wait for either motion or button.
function waitForMotion()
	print("AP configured.  Waiting for motion or button...")
	gpio.trig(CONFIG_BTN, "down", fireconfig)
	gpio.trig(CONFIG_MOTION, "up", firemotion)
end

-- Unconfigured state.  Wait for button only.
function waitForConfig()
	print("No AP configuration yet.  Waiting for buttonpress...")
	gpio.trig(CONFIG_BTN, "down", fireconfig)
end

-- Turn off all the timers and GPIO triggers.
function reset()
	-- Make sure any extant triggers are off
	gpio.mode(CONFIG_BTN, gpio.INPUT)
	gpio.mode(CONFIG_MOTION, gpio.INPUT)
	gpio.trig(CONFIG_BTN)
	gpio.trig(CONFIG_MOTION)
	tmr.unregister(TIMER_SETUP)
	tmr.unregister(TIMER_TEARDOWN)
	wifi.sta.eventMonStop()
end

-- First state from boot time.  Figure out (based on whether the AP has ever
-- been configured) which wait state to enter.
function startup()
	print("Startup")
	-- Turn the radio off on boot.
	wifi.nullmodesleep(false)
	wifi.setmode(wifi.NULLMODE, true)
	-- Turn off the triggers
	reset()
	-- Check to see if we have an AP config already and branch to either
	-- the WAITING state or the UNCONFIGURED state
	conf = wifi.sta.getconfig()
	if conf == "" then
		-- Nothing set yet
		waitForConfig()
	else
		-- An AP has been set
		waitForMotion()
	end
end

-- Enter state 1 (startup)
startup()
