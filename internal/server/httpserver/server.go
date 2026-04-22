package httpserver

type Server struct {
	AuthServer
	TemplatesServer
	HistoryServer
	DictServer
	AIQueryServer
}

func NewServer(authServer AuthServer, templatesServer TemplatesServer, historyServer HistoryServer, dictServer DictServer, aiQueryServer AIQueryServer) *Server {
	var h = &Server{
		AuthServer:      authServer,
		TemplatesServer: templatesServer,
		HistoryServer:   historyServer,
		DictServer:      dictServer,
		AIQueryServer:   aiQueryServer,
	}

	return h
}
