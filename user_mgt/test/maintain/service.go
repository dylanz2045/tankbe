package maintain

import (
	"testing"
	"user_mgt/user_mgt/maintain"
)

type ServiceTest interface {
	TestGetOnlineUsersAmount(t *testing.T) (amount int)
}

type ServiceTestImpl struct {
}

func NewServiceTest() ServiceTest {
	return &ServiceTestImpl{}
}
func (*ServiceTestImpl) TestGetOnlineUsersAmount(t *testing.T) (amount int) {
	maintainer := maintain.NewOnlineUser()

	amount, err := maintainer.GetOnlineUserAmount()
	if err != nil {
		t.Fatalf("failed to get online users amount: %s", err)
	}

	t.Logf("get online users amount success: %d", amount)

	return amount
}
