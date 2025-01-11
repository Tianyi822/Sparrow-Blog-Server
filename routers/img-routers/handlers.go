package img_routers

import (
	"h2blog/pkg/resp"
	"h2blog/pkg/utils"
	"h2blog/pkg/webp"
	"h2blog/routers/tools"
	"time"

	"github.com/gin-gonic/gin"
)

func uploadImages(ctx *gin.Context) {
	// Get imgs data from raw data
	imgsDto, err := tools.GetImgsDto(ctx)
	if err != nil {
		return
	}

	resp.Ok(ctx, "上传成功", imgsDto)
}

func streamConvertProgress(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	tracker := webp.Converter.GetProgress()
	if tracker == nil {
		resp.BadRequest(c, "没有正在进行的转换任务", nil)
		return
	}

	// 生成客户端ID
	clientID := utils.GenUUID()

	// 订阅进度更新
	ch := tracker.Subscribe(clientID)
	defer tracker.Unsubscribe(clientID)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.Request.Context().Done():
				return
			case imgDto := <-ch:
				// 处理单个任务完成事件
				c.SSEvent("task", imgDto)
				c.Writer.Flush()
			case <-ticker.C:
				// 定期发送整体进度
				total, success, failed := tracker.GetProgress()
				c.SSEvent("progress", gin.H{
					"total":    total,
					"success":  success,
					"failed":   failed,
					"finished": success+failed >= total,
				})
				c.Writer.Flush()

				if success+failed >= total {
					return
				}
			}
		}
	}()

	<-c.Request.Context().Done()
}
