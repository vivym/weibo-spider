package nlp

import (
	"github.com/vivym/weibo-spider/internal/api/protobuf/nlp"
	"google.golang.org/grpc"
)

type NLPToolkit struct {
	conn   *grpc.ClientConn
	client nlp.NLPClient
}

func New(config Config) (*NLPToolkit, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(config.Address, opts...)
	if err != nil {
		return nil, err
	}

	client := nlp.NewNLPClient(conn)

	return &NLPToolkit{
		conn:   conn,
		client: client,
	}, nil
}

func (n *NLPToolkit) Release() {
	if n.conn != nil {
		_ = n.conn.Close()
	}
}
