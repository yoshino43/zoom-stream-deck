package main

import (
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"meow.tf/streamdeck/sdk"
	"os/exec"
	"strings"
)

func main() {
	sdk.RegisterAction("yy.zoom.controller.mute", MuteAction)
	sdk.RegisterAction("yy.zoom.controller.video", MuteAction)
	sdk.RegisterAction("yy.zoom.controller.share", MuteAction)
	sdk.RegisterAction("yy.zoom.controller.leave", MuteAction)
	sdk.RegisterAction("yy.zoom.controller.record", MuteAction)
	// Open and connect the SDK
	if err := sdk.Open(); err != nil {
		log.Fatalln(err)
	}
	sdk.Log("yy.zoom.controller started")
	// Wait until the socket is closed, or SIGTERM/SIGINT is received
	sdk.Wait()
}

type ZoomStatus struct {
	muteStatus   string
	videoStatus  string
	shareStatus  string
	recordStatus string
}

func GetZoomStatus() (*ZoomStatus, error) {
	var zoomStatus = `set zoomStatus to "closed"
set muteStatus to "disabled"
set videoStatus to "disabled"
set shareStatus to "disabled"
set recordStatus to "disabled"
tell application "System Events"
	if exists (window 1 of process "zoom.us") then
		set zoomStatus to "open"
		tell application process "zoom.us"
			if exists (menu bar item "ミーティング" of menu bar 1) then
				set zoomStatus to "call"
				if exists (menu item "オーディオのミュート" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
					set muteStatus to "unmuted"
				else
					set muteStatus to "muted"
				end if
				if exists (menu item "ビデオの開始" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
					set videoStatus to "stopped"
				else
					set videoStatus to "started"
				end if
			end if
		end tell
	end if
end tell

do shell script "echo zoomMute:" & (muteStatus as text) & ",zoomVideo:" & (videoStatus as text) & ",zoomStatus:" & (zoomStatus as text) & ",zoomShare:" & (shareStatus as text) & ",zoomRecord:" & (recordStatus as text)
`
	out, err := exec.Command("osascript", "-e", zoomStatus).Output()
	if err != nil {
		return nil, err
	}
	status := &ZoomStatus{}
	for _, s := range strings.Split(string(out), ",") {
		ss := strings.Split(s, ":")
		if len(ss) < 2 {
			continue
		}
		switch ss[0] {
		case "muteStatus":
			status.muteStatus = ss[1]
		case "videoStatus":
			status.videoStatus = ss[1]
		case "shareStatus":
			status.shareStatus = ss[1]
		case "recordStatus":
			status.recordStatus = ss[1]
		}
	}
	return status, nil

}

func MuteAction(action, context string, payload *fastjson.Value, deviceId string) {
	sdk.Log(fmt.Sprintf("Action %s context %s payload %#v device %s\n", action, context, payload, deviceId))
	//status, err := GetZoomStatus()
	//if err != nil {
	//	sdk.Log(err.Error())
	//}
	var cmd = `tell application "zoom.us"
	tell application "System Events" to tell application process "zoom.us"
		if exists (menu item "オーディオのミュート" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
			click (menu item "オーディオのミュート" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		else
			click (menu item "オーディオのミュート解除" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		end if
	end tell
end tell
`
	if err := exec.Command("osascript", "-e", cmd).Run(); err != nil {
		sdk.Log(err.Error())
	}
}
