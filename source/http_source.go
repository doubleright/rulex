package source

import (
	"context"
	"fmt"
	"net/http"
	"rulex/core"
	"rulex/typex"
	"rulex/utils"

	"github.com/gin-gonic/gin"
	"github.com/ngaut/log"
)

//
type httpConfig struct {
	Port       uint16             `json:"port" validate:"required" title:"端口" info:""`
	DataModels []typex.XDataModel `json:"dataModels" title:"数据模型" info:""`
}

//
type httpInEndSource struct {
	typex.XStatus
	engine *gin.Engine
}

func NewHttpInEndSource(inEndId string, e typex.RuleX) typex.XSource {
	h := httpInEndSource{}
	h.PointId = inEndId
	gin.SetMode(gin.ReleaseMode)
	h.engine = gin.New()
	h.RuleEngine = e
	return &h
}
func (*httpInEndSource) Configs() *typex.XConfig {
	return core.GenInConfig(typex.HTTP, "HTTP", httpConfig{})
}

//
func (hh *httpInEndSource) Start() error {
	config := hh.RuleEngine.GetInEnd(hh.PointId).Config
	var mainConfig httpConfig
	if err := utils.BindSourceConfig(config, &mainConfig); err != nil {
		return err
	}
	hh.engine.POST("/in", func(c *gin.Context) {
		type Form struct {
			Data string
		}
		var inForm Form
		err := c.BindJSON(&inForm)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			hh.RuleEngine.Work(hh.RuleEngine.GetInEnd(hh.PointId), inForm.Data)
			c.JSON(200, gin.H{
				"message": "ok",
				"data":    inForm,
			})
		}
	})
	ctx, cancelCTX := context.WithCancel(typex.GCTX)
	hh.Ctx = ctx
	hh.CancelCTX = cancelCTX

	go func(ctx context.Context) {
		http.ListenAndServe(fmt.Sprintf(":%v", mainConfig.Port), hh.engine)
	}(ctx)
	log.Info("HTTP source started on" + " [0.0.0.0]:" + fmt.Sprintf("%v", mainConfig.Port))

	return nil
}

//
func (mm *httpInEndSource) DataModels() []typex.XDataModel {
	return []typex.XDataModel{}
}

//
func (hh *httpInEndSource) Stop() {
	hh.CancelCTX()

}
func (hh *httpInEndSource) Reload() {

}
func (hh *httpInEndSource) Pause() {

}
func (hh *httpInEndSource) Status() typex.SourceState {
	return typex.UP
}

func (hh *httpInEndSource) Register(inEndId string) error {
	hh.PointId = inEndId
	return nil
}

func (hh *httpInEndSource) Test(inEndId string) bool {
	return true
}

func (hh *httpInEndSource) Enabled() bool {
	return hh.Enable
}
func (hh *httpInEndSource) Details() *typex.InEnd {
	return hh.RuleEngine.GetInEnd(hh.PointId)
}
func (m *httpInEndSource) OnStreamApproached(data string) error {
	return nil
}
func (*httpInEndSource) Driver() typex.XExternalDriver {
	return nil
}

//
// 拓扑
//
func (*httpInEndSource) Topology() []typex.TopologyPoint {
	return []typex.TopologyPoint{}
}
