package webserver

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/contact"
	"github.com/nanu-c/textsecure-qml/app/helpers"
	"github.com/nanu-c/textsecure-qml/app/store"
)

func sendChatList(client *websocket.Conn) {
	var err error
	chatListEnvelope := &ChatListEnvelope{
		ChatList: store.SessionsModel.Sess,
	}
	message := &[]byte{}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.WriteMessage(websocket.TextMessage, *message)
	if err != nil {
		log.Println("[textsecure] send error chat list ", err)
		return
	}
}
func sendCurrentChat(client *websocket.Conn, s *store.Session) {
	var err error
	currentChatEnvelope := &CurrentChatEnvelope{
		CurrentChat: s,
	}
	message := &[]byte{}
	*message, err = json.Marshal(currentChatEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.WriteMessage(websocket.TextMessage, *message)
	if err != nil {
		log.Println("[textsecure] send error current chat ", err)
		return
	}

}
func refreshContacts(client *websocket.Conn, path string) {
	var err error
	config.VcardPath = path
	contact.GetAddressBookContactsFromContentHub()
	err = store.RefreshContacts()
	if err != nil {
		ShowError(err.Error())
	}
	// go sendContactList(client)
}
func sendContactList(client *websocket.Conn) {
	var err error
	contactListEnvelope := &ContactListEnvelope{
		ContactList: store.ContactsModel.Contacts,
	}
	message := &[]byte{}
	*message, err = json.Marshal(contactListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send error contact list ", err)
		return
	}
}
func sendDeviceList(client *websocket.Conn) {
	var err error
	devices, err := textsecure.LinkedDevices()
	deviceListEnvelope := &DeviceListEnvelope{
		DeviceList: devices,
	}
	message := &[]byte{}
	*message, err = json.Marshal(deviceListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send error device list", err)
		return
	}
}
func createChat(tel string) *store.Session {
	return store.SessionsModel.Get(tel)
}
func createGroup(newGroupData CreateGroupMessage) *store.Session {
	group, err := textsecure.NewGroup(newGroupData.Name, newGroupData.Members)
	if err != nil {
		ShowError(err.Error())
		return nil
	}
	members := strings.Join(newGroupData.Members, ",")
	if !strings.Contains(members, config.Config.Tel) {
		// log.Debugln(members, config.Config.Tel)
		members = members + "," + config.Config.Tel
	}
	store.Groups[group.Hexid] = &store.GroupRecord{
		GroupID: group.Hexid,
		Name:    newGroupData.Name,
		Members: members,
	}
	store.SaveGroup(store.Groups[group.Hexid])
	session := store.SessionsModel.Get(group.Hexid)
	msg := session.Add(store.GroupUpdateMsg(append(newGroupData.Members, config.Config.Tel), newGroupData.Name), "", "", "", true, store.ActiveSessionID)
	msg.Flags = helpers.MsgFlagGroupNew
	//qml.Changed(msg, &msg.Flags)
	store.SaveMessage(msg)

	return session
}
func sendMessageList(client *websocket.Conn, id string) {
	message := &[]byte{}

	err, messageList := store.SessionsModel.GetMessageList(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	messageList.Session.MarkRead()
	chatListEnvelope := &MessageListEnvelope{
		MessageList: messageList,
	}
	*message, err = json.Marshal(chatListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send error message list", err)
		return
	}
}
func sendMoreMessageList(client *websocket.Conn, id string, lastId string) {
	message := &[]byte{}
	err, messageList := store.SessionsModel.GetMoreMessageList(id, lastId)
	if err != nil {
		fmt.Println(err)
		return
	}
	moreMessageListEnvelope := &MoreMessageListEnvelope{
		MoreMessageList: messageList,
	}
	*message, err = json.Marshal(moreMessageListEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send error more message list", err)
		return
	}
}
func sendIdentityInfo(client *websocket.Conn, myId []byte, theirId []byte) {
	var err error

	message := &[]byte{}
	identityEnvelope := &IdentityEnvelope{
		Identity: fmt.Sprintf("% 0X", myId),
		TheirId:  fmt.Sprintf("% 0X", theirId),
	}
	*message, err = json.Marshal(identityEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send error identity info", err)
		return
	}

}

var test = false

func UpdateChatList() {

	if activeChat == "" {
		for client := range clients {
			sendChatList(client)
		}
	}
}
func UpdateContactList() {

	if activeChat == "" {
		for client := range clients {
			sendContactList(client)
		}
	}
}

type SendGui struct {
	Gui string
}

func SetGui() {

	if activeChat == "" {
		for client := range clients {
			var err error
			gui := config.Gui
			if gui == "" {
				gui = "standard"
			}
			request := &SendGui{
				Gui: gui,
			}
			message := &[]byte{}
			*message, err = json.Marshal(request)
			if err != nil {
				log.Errorln("[axolotl] set gui", err)
				return
			}
			if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
				log.Println("[textsecure] send error set gui", err)
				return
			}
		}
	}
}
func sendConfig(client *websocket.Conn) {
	var err error

	message := &[]byte{}
	configEnvelope := &ConfigEnvelope{
		Type:             "config",
		RegisteredNumber: config.Config.Tel,
		Version:          config.AppVersion,
	}
	*message, err = json.Marshal(configEnvelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.WriteMessage(websocket.TextMessage, *message); err != nil {
		log.Println("[textsecure] send config", err)
		return
	}
}
