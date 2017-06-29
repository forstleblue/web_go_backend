package render

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

//JSON renders and sends UTF8 application/json data to the client
func JSON(ctx *fasthttp.RequestCtx, data interface{}, message string, HTTPStatus int) {

	ctx.SetContentType("application/json; charset=utf-8")

	var jsData, jsMessage []byte
	var err error

	if data != nil {
		if jsData, err = json.Marshal(data); err != nil {
			//renderError(ctx, "Error in json.Marsha(): "+err.Error())
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBody([]byte("{\"data\":\"Error\",\"message\":\"Could not marshal data to JSON\"}"))
			return
		}
	}
	if jsMessage, err = json.Marshal(message); err != nil {
		//renderError(ctx, "Error in json.Marsha(): "+err.Error())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte("{\"data\":\"Error\",\"message\":\"Could not marshal message to JSON\"}"))
		return
	}

	ctx.SetStatusCode(HTTPStatus)
	if data == nil {
		ctx.SetBody([]byte(`{"message":` + string(jsMessage) + `}`))
	} else {
		ctx.SetBody([]byte(`{"data":` + string(jsData) + `,"message":` + string(jsMessage) + `}`))
	}
}
