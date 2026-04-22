package httpserver

type Server struct {
	AuthServer
	TemplatesServer
	HistoryServer
	DictServer
	ExecutorServer
}

func NewServer(authServer AuthServer, templatesServer TemplatesServer, historyServer HistoryServer, dictServer DictServer, executorServer ExecutorServer) *Server {
	var h = &Server{
		AuthServer:      authServer,
		TemplatesServer: templatesServer,
		HistoryServer:   historyServer,
		DictServer:      dictServer,
		ExecutorServer:  executorServer,
	}

	return h
}
