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

-- Create the timer we'll use to tear down the wifi after sending the barks
teardowntimer = tmr.create()
setuptimer = tmr.create()

-- Turn the radio off on boot.
wifi.nullmodesleep(false)
wifi.setmode(wifi.NULLMODE, true)

-- Set up an edge trigger to respond to the pushbutton (for smartconfig).
-- We set the GPIO trigger back up at the end for the next button press.
function setconfig(ssid, password)
	setuptimer:stop()
	wifi.stopsmart()
	config = {}
	config.ssid = ssid
	config.pwd = password
	config.auto = true
	config.save = true
	wifi.sta.config(config)
	gpio.trig(CONFIG_BTN, "down", fireconfig)
end

-- Callback when we get a config pushbutton down.
function fireconfig(level, when)
	gpio.trig(CONFIG_BTN, "none")
	wifi.setmode(wifi.STATION)
	setuptimer:register(CONFIG_SETUPTIME, tmr.ALARM_SINGLE, setuptimeout)
	setuptimer:start()
	wifi.startsmart(0, setconfig)
end

-- Callback when setup has been running without success for too long.
function setuptimeout()
	setuptimer:stop()
	wifi.stopsmart()
	gpio.trig(CONFIG_BTN, "down", fireconfig)
end

-- Callback for the motion sensor.
function firemotion(level, when)
	wifi.setmode(wifi.STATION, false)
	wifi.eventMonReg(wifi.STA_GOTIP, sendmotion)
end

-- Callback for sending the motion sensor packet once we get an IP.
function sendmotion(previous_state)
	-- Figure out what the broadcast address is to send the packet.
	bip = wifi.sta.getbroadcast()
	sock = net.createUDPSocket()
	sock:send(CONFIG_PORT, bip, onpacket)
	sock:send(CONFIG_PORT, bip, onpacket)
	if teardowntimer ~= nil then
		teardowntimer:register(CONFIG_WIFITIME, tmr.ALARM_SINGLE, teardown)
		teardowntimer:start()
	end
	gpio.trig(CONFIG_MOTION, "down", firemotion)
end

-- Tear down the Wifi after the packets are sent
function teardown()
	teardowntimer:stop()
	wifi.setmode(wifi.NULLMODE, false)
	gpio.trig(CONFIG_MOTION, "down", firemotion)
end

-- Finally, install the gpio handlers and enter the main loop.
gpio.trig(CONFIG_BTN, "down", fireconfig)
gpio.trig(CONFIG_MOTION, "down", firemotion)
