package notifications

import (
	pb "01cloud-api/api/notifications/proto"
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func NotifyEmail(conn grpc.ClientConnInterface, data *SendEmailMessage) error {
	users := data.User
	for _, user := range users {
		mes := &pb.EmailMessage{}
		data.User = []string{user}
		data.ToGrpc(mes)
		c := pb.NewNotificationClient(conn.(*grpc.ClientConn))
		_, err := c.SendEmail(context.Background(), &pb.SendEmailRequest{Message: mes})
		if err != nil {
			logrus.Error(err)
			continue
		}
	}
	return nil

}
