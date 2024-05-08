package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Dtos"
	"server/helper"
	"server/internal/ws"
	"server/ports"
)

type Handler struct {
	hub         *ws.Hub
	roomService ports.RoomServicePort
	wrapper     *helper.ResponserWrapper
}

func NewHandler(hub *ws.Hub, roomService ports.RoomServicePort) *Handler {
	return &Handler{roomService: roomService, hub: hub, wrapper: helper.NewResponseWrapper()}
}

func (receiver Handler) FindRoom(c *gin.Context) {
	//var Request Dtos.GetAllRoomFilterDto

	room := receiver.roomService.RetrieveRoom(c.Param("id"))
	receiver.wrapper.SetResource(room.ToTransformer())
	c.JSON(200, receiver.wrapper.GetResource())

}

func (receiver Handler) FetchRooms(c *gin.Context) {
	dto := new(Dtos.GetAllRoomFilterDto)
	rooms := receiver.roomService.FetchAllRooms(1, 10, *dto)

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
