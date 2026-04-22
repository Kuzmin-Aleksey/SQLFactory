package httpserver

type Server struct {
	AuthServer
}

func NewServer(authServer AuthServer) *Server {
	var h = &Server{
		AuthServer: authServer,
	}

	return h
}
