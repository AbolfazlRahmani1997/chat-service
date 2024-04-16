package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Dtos"
	"server/internal/ws"
	"server/ports"
)

type Handler struct {
	hub         *ws.Hub
	roomService ports.RoomServicePort
}

func NewHandler(hub *ws.Hub, roomService ports.RoomServicePort) Handler {
	return Handler{roomService: roomService, hub: hub}
}

func (receiver Handler) FindRoom(c *gin.Context) {

	roomId := c.Param("roomId")
	room := receiver.roomService.RetrieveRoom(roomId)
	c.JSON(200, room)

}

func (receiver Handler) FetchRooms(c *gin.Context) {
	dto := new(Dtos.GetAllRoomFilterDto)

	err := c.BindJSON(dto)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rooms := receiver.roomService.FetchAllRooms(*dto)
	c.JSON(200, rooms)
}

func (receiver Handler) EditRoom(c *gin.Context) {
	dto := new(Dtos.UpdateRoomDto)

	err := c.BindJSON(dto)
	if err != nil {
		fmt.Println("z")

		return
	}
	dto.Id = c.Param("roomId")
	fmt.Println(dto.Id)
	rooms := receiver.roomService.EditRooms(*dto)
	c.JSON(200, rooms)
}
