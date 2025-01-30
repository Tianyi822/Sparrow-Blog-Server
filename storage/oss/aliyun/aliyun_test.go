package aliyun

import (
	"fmt"
	"github.com/aliyun/credentials-go/credentials"
	"testing"
)

func TestKey2(t *testing.T) {
	config := new(credentials.Config).
		// 设置凭证类型
		SetType("ram_role_arn").
		// 用户 AccessKey Id
		SetAccessKeyId("LTAI5tBna5CmNYTqW5Fd7hJv").
		// 用户 AccessKey Secret
		SetAccessKeySecret("2AiDSaYcoqAE8AVVvak2rdI3xt0ljN").
		// 要扮演的RAM角色ARN，示例值：acs:ram::123456789012****:role/adminrole，可以通过环境变量ALIBABA_CLOUD_ROLE_ARN设置RoleArn
		SetRoleArn("acs:ram::1342607901578627:role/ramosstest").
		// 角色会话名称，可以通过环境变量ALIBABA_CLOUD_ROLE_SESSION_NAME设置RoleSessionName
		SetRoleSessionName("RamOssTest").
		// 角色会话过期时间，单位秒，默认3600秒，最大3600秒
		SetRoleSessionExpiration(3600).
		// 设置 SessionName
		SetRoleSessionName("RamOssTest")

	provider, err := credentials.NewCredential(config)
	if err != nil {
		return
	}
	credential, err := provider.GetCredential()
	if err != nil {
		return
	}

	fmt.Printf("%#v", credential)
}
