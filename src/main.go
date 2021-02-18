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
	sdk.RegisterAction("com.yy.zoom.controller.mute", MuteAction)
	sdk.RegisterAction("com.yy.zoom.controller.video", VideoAction)
	sdk.RegisterAction("com.yy.zoom.controller.share", ShareAction)
	sdk.RegisterAction("com.yy.zoom.controller.leave", LeaveAction)
	sdk.RegisterAction("com.yy.zoom.controller.record", RecordAction)
	// Open and connect the SDK
	if err := sdk.Open(); err != nil {
		log.Fatalln(err)
	}
	sdk.Log("com.yy.zoom.controller started")
	// Wait until the socket is closed, or SIGTERM/SIGINT is received
	sdk.Wait()
}

const (
	Inactive = iota
	Active
)

type ZoomStatus struct {
	ZoomEnabled  bool
	muteStatus   string
	videoStatus  string
	shareStatus  string
	recordStatus string
}

func (s *ZoomStatus) isMute() bool {
	switch s.muteStatus {
	case "disabled", "muted":
		return true
	default:
		return false
	}
}

func (s *ZoomStatus) isActiveVideo() bool {
	switch s.videoStatus {
	case "started":
		return true
	default:
		return false
	}
}

func (s *ZoomStatus) isActiveRecording() bool {
	switch s.recordStatus {
	case "started":
		return true
	default:
		return false
	}
}

func GetZoomStatus() (*ZoomStatus, error) {
	var cmd = `
set zoomStatus to "closed"
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
				if exists (menu item "共有の開始" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
					set shareStatus to "stopped"
				else
					set shareStatus to "started"
				end if
				if exists (menu item "レコーディング" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
					set recordStatus to "stopped"
				else if exists (menu item "レコーディングを再開" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
					set recordStatus to "stopped"
				else
					set recordStatus to "started"
				end if
			end if
		end tell
	end if
end tell

do shell script "echo muteStatus:" & (muteStatus as text) & ",videoStatus:" & (videoStatus as text) & ",zoomStatus:" & (zoomStatus as text) & ",shareStatus:" & (shareStatus as text) & ",recordStatus:" & (recordStatus as text)
`
	out, err := exec.Command("osascript", "-e", cmd).Output()
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
		case "zoomStatus":
			status.ZoomEnabled = ss[1] == "call"
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
	before, _ := GetZoomStatus()
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
	if before != nil {
		if before.isMute() {
			sdk.SetState(context, Active)
		} else {
			sdk.SetState(context, Inactive)
		}

	}

}

func VideoAction(action, context string, payload *fastjson.Value, deviceId string) {
	sdk.Log(fmt.Sprintf("Action %s context %s payload %#v device %s\n", action, context, payload, deviceId))
	before, _ := GetZoomStatus()

	var cmd = `tell application "zoom.us"
	tell application "System Events" to tell application process "zoom.us"
		if exists (menu item "ビデオの開始" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
			click (menu item "ビデオの開始" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		else
			click (menu item "ビデオの停止" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		end if
	end tell
end tell
`
	if err := exec.Command("osascript", "-e", cmd).Run(); err != nil {
		sdk.Log(err.Error())
	}

	if before != nil {
		if before.isActiveVideo() {
			sdk.SetState(context, Inactive)
		} else {
			sdk.SetState(context, Active)
		}

	}
}

func ShareAction(action, context string, payload *fastjson.Value, deviceId string) {
	sdk.Log(fmt.Sprintf("Action %s context %s payload %#v device %s\n", action, context, payload, deviceId))

	var cmd = `tell application "zoom.us"
	tell application "System Events" to tell application process "zoom.us"
		if exists (menu item "共有の開始" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
			click (menu item "共有の開始" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		else
			click (menu item "共有の停止" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		end if
	end tell
end tell
`
	if err := exec.Command("osascript", "-e", cmd).Run(); err != nil {
		sdk.Log(err.Error())
	}
}

func LeaveAction(action, context string, payload *fastjson.Value, deviceId string) {
	sdk.Log(fmt.Sprintf("Action %s context %s payload %#v device %s\n", action, context, payload, deviceId))
	var cmd = `
tell application "System Events"
	if exists (menu bar item "ミーティング" of menu bar 1 of application process "zoom.us") then
		tell application "zoom.us" to activate
		keystroke "w" using {command down}
		tell front window of (first application process whose frontmost is true)
			click button 1
		end tell
	end if
end tell

`
	if err := exec.Command("osascript", "-e", cmd).Run(); err != nil {
		sdk.Log(err.Error())
	}
	sdk.SetState(context, Inactive)
}

func RecordAction(action, context string, payload *fastjson.Value, deviceId string) {
	sdk.Log(fmt.Sprintf("Action %s context %s payload %#v device %s\n", action, context, payload, deviceId))
	before, _ := GetZoomStatus()
	var cmd = `tell application "zoom.us"
	tell application "System Events" to tell application process "zoom.us"
		if exists (menu item "レコーディング" of menu 1 of menu bar item "ミーティング" of menu bar 1) then
			click (menu item "レコーディング" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		else
			click (menu item "レコーディングを停止" of menu 1 of menu bar item "ミーティング" of menu bar 1)
		end if
	end tell
end tell
`
	if err := exec.Command("osascript", "-e", cmd).Run(); err != nil {
		sdk.Log(err.Error())
	}

	if before != nil {
		if before.isActiveRecording() {
			sdk.SetState(context, Inactive)
		} else {
			sdk.SetState(context, Active)
		}
	}
}
