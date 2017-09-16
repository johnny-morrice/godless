package http

import (
	"net/http"

	"github.com/johnny-morrice/godless/api"
)

type ProfilingWebService struct {
	ProfilingWebServiceOptions
	wsHandler http.Handler

	timerList *api.TimerList
}

type ProfilingWebServiceOptions struct {
	WS   api.WebService
	Prof api.Profiler
}

func MakeProfilingWebService(options ProfilingWebServiceOptions) api.WebService {
	if options.WS == nil {
		options.WS = MakeWebService(api.WebServiceOptions{})
	}

	if options.Prof == nil {
		panic("options.Prof was nil")
	}

	return &ProfilingWebService{
		ProfilingWebServiceOptions: options,
		wsHandler:                  options.WS.GetApiRequestHandler(),
		timerList:                  &api.TimerList{},
	}
}

func (service *ProfilingWebService) SetOptions(options api.WebServiceOptions) {
	service.WS.SetOptions(options)
}

func (service *ProfilingWebService) GetApiRequestHandler() http.Handler {
	return http.HandlerFunc(service.handleApiRequest)
}

func (service *ProfilingWebService) handleApiRequest(rw http.ResponseWriter, req *http.Request) {
	const profileName = __PROFILE_WEB_SERVICE_NAME + ".handleApiRequest"

	timer := service.Prof.NewTimer(profileName)

	service.timerList.StartTimer(timer)
	defer service.timerList.StopTimer(timer)

	service.wsHandler.ServeHTTP(rw, req)
}

func (service *ProfilingWebService) Close() {
	defer service.timerList.StopAllTimers()
	service.WS.Close()
}

const __PROFILE_WEB_SERVICE_NAME = "ProfilingWebService"
