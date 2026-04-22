package httpserver

type Server struct {
	AuthServer
	TemplatesServer
	HistoryServer
	DictServer
}

func NewServer(authServer AuthServer, templatesServer TemplatesServer, historyServer HistoryServer, dictServer DictServer) *Server {
	var h = &Server{
		AuthServer:      authServer,
		TemplatesServer: templatesServer,
		HistoryServer:   historyServer,
		DictServer:      dictServer,
	}

	return h
}
