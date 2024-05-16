package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Dtos"
	"server/entity/Room"
	"server/helper"
	"server/internal/ws"
	"server/ports/services"
)

type Handler struct {
	hub            *ws.Hub
	roomService    services.RoomServicePort
	messageService services.MessageServicePort
	wrapper        *helper.ResponserWrapper
}

func NewHandler(hub *ws.Hub, roomService services.RoomServicePort, message services.MessageServicePort) *Handler {
	return &Handler{roomService: roomService, messageService: message, hub: hub, wrapper: helper.NewResponseWrapper()}
}

func (receiver Handler) FindRoom(c *gin.Context) {

	room := receiver.roomService.RetrieveRoom(c.Param("roomId"))
	receiver.wrapper.SetResource(&room)
	c.JSON(200, receiver.wrapper.GetResource())

}

func (receiver Handler) FetchRooms(c *gin.Context) {
	dto := &Dtos.GetAllRoomFilterDto{
		// Fill DTO fields using query parameters
		// Example:
		MemberId: c.Query("userId"),
		// Adjust these field assignments based on your actual DTO structure
	}
	var rooms []Room.Room
	rooms = receiver.roomService.FetchAllRooms(1, 10, *dto)
	var data []interface{}
	for _, room := range rooms {
		data = append(data, room)
	}
	receiver.wrapper.SetCollectionResource(data)
	c.JSON(200, rooms)
}

func (receiver Handler) EditRoom(c *gin.Context) {
	dto := new(Dtos.UpdateRoomDto)

	err := c.BindJSON(dto)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dto.Id = c.Param("roomId")
	fmt.Println(dto.Id)
	rooms := receiver.roomService.EditRooms(*dto)
	c.JSON(200, rooms)
}

func (receiver Handler) GetMessage(c *gin.Context) {

	message := receiver.messageService.GetMessage(c.Query("id"))
	c.JSON(200, message)
}
func (receiver Handler) GetAllMessages(c *gin.Context) {

	message := receiver.messageService.GetMessage(c.Query("id"))
	c.JSON(200, message)
}
