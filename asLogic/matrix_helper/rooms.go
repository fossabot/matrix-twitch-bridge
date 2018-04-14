package matrix_helper

import "github.com/matrix-org/gomatrix"

func CreateRoom(client *gomatrix.Client, displayname, avatarURL, alias, preset string) (*gomatrix.RespCreateRoom, error) {
	createRoomReq := &gomatrix.ReqCreateRoom{}
	createRoomReq.Name = displayname

	if avatarURL != "" {
		resp, err := client.UploadLink(avatarURL)
		if err != nil {
			return nil, err
		}
		content := make(map[string]interface{})
		content["url"] = resp.ContentURI
		m_room_avatar_event := gomatrix.Event{
			Type:    "m.room.avatar",
			Content: content,
		}
		createRoomReq.InitialState = append(createRoomReq.InitialState, m_room_avatar_event)
	}

	if alias != "" {
		createRoomReq.RoomAliasName = alias
	}

	createRoomReq.Preset = preset
	roomResp, err := client.CreateRoom(createRoomReq)
	if err != nil {
		return nil, err
	}
	return roomResp, nil
}
