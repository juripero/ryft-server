package rest

import (
	"net/http"
	"sort"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetFileParams query parameters for GET /files
type GetFilesParams struct {
	Dir    string `form:"dir" json:"dir"`       // directory to get content of
	Hidden bool   `form:"hidden" json:"hidden"` // show hidden files/dirs
	Local  bool   `form:"local" json:"local"`
}

// GET /files method
func (server *Server) DoGetFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := GetFilesParams{}
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
		// log.Debugf("[%s]: Content-Type changed to %s", CORE, accept)
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewError(http.StatusUnsupportedMediaType,
			"only JSON format is supported for now"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err := server.getSearchEngine(params.Local, nil /*no files*/, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	log.WithFields(map[string]interface{}{
		"dir":     params.Dir,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /files", CORE)
	info, err := engine.Files(params.Dir, params.Hidden)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get files"))
	}

	// TODO: if params.Sort {
	// sort names in the ascending order
	sort.Strings(info.Files)
	sort.Strings(info.Dirs)

	// TODO: use transcoder/dedicated structure instead of simple map!
	json := map[string]interface{}{
		"dir":     info.Path,
		"files":   info.Files,
		"folders": info.Dirs,
	}
	ctx.JSON(http.StatusOK, json)
}
