package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/VillanCh/ragsync/common/aliyun"
	"github.com/davecgh/go-spew/spew"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

func main() {
	app := cli.NewApp()
	app.Name = "ragsync"
	app.Usage = "Aliyun Bailian RAG Sync Tool"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "阿里云 AccessKey ID",
			EnvVar: "ALIYUN_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "secret",
			Usage:  "阿里云 AccessKey Secret",
			EnvVar: "ALIYUN_SECRET",
		},
		cli.StringFlag{
			Name:  "endpoint",
			Usage: "百炼服务端点",
			Value: "bailian.cn-beijing.aliyuncs.com",
		},
		cli.StringFlag{
			Name:   "workspace-id",
			Usage:  "百炼工作空间ID",
			EnvVar: "BAILIAN_WORKSPACE_ID",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "apply-lease",
			Usage: "申请文件上传租约",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "要上传的文件路径",
				},
			},
			Action: func(c *cli.Context) error {
				client, err := aliyun.NewBailianClient(
					c.GlobalString("access-key"),
					c.GlobalString("secret"),
					c.GlobalString("endpoint"),
				)
				if err != nil {
					return err
				}

				workspaceId := c.GlobalString("workspace-id")
				if workspaceId == "" {
					return utils.Errorf("请指定百炼工作空间ID")
				}
				client.SetWorkspaceId(workspaceId)

				lis, err := client.ApplyFileUploadLease("test.txt", []byte("test"))
				if err != nil {
					return err
				}
				spew.Dump(lis)

				headers := utils.InterfaceToGeneralMap(lis.Headers)
				bailianExtra, ok := headers["X-bailian-extra"]
				if !ok {
					return utils.Errorf("X-bailian-extra 不存在")
				}
				contentType, ok := headers["Content-Type"]
				if !ok {
					return utils.Errorf("Content-Type 不存在")
				}

				// 上传文件
				content := []byte("test")
				err = aliyun.UploadFile(lis.Method, lis.UploadURL, "test.txt", fmt.Sprint(contentType), content, fmt.Sprintf("%s", bailianExtra))
				if err != nil {
					return err
				}

				log.Info("添加文件到百炼 RAG")
				err = client.AddFile(lis.LeaseId)
				if err != nil {
					return err
				}

				log.Info("文件添加成功")
				return nil
			},
		},
		{
			Name:  "upload",
			Usage: "上传文件到百炼 RAG",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "要上传的文件路径",
				},
			},
			Action: func(c *cli.Context) error {
				agentName := c.String("agent-name")
				filePath := c.String("file")
				if filePath == "" {
					return utils.Errorf("请指定要上传的文件路径")
				}
				// 创建百炼客户端
				client, err := aliyun.NewBailianClient(
					c.GlobalString("access-key"),
					c.GlobalString("secret"),
					c.GlobalString("endpoint"),
				)
				if err != nil {
					return err
				}

				// 申请文件上传租约
				lease, err := client.ApplyFileUploadLease(agentName, []byte("test"))
				if err != nil {
					return err
				}
				log.Infof("成功获取上传租约，租约ID: %s", lease.LeaseId)

				// headers := utils.InterfaceToGeneralMap(lease.Headers)
				// bailianExtra, ok := headers["X-bailian-extra"]
				// if !ok {
				// 	return utils.Errorf("X-bailian-extra 不存在")
				// }

				// // 上传文件
				// content := []byte("test")
				// err = aliyun.UploadFile(lease.Method, lease.UploadURL, filePath, content, fmt.Sprintf("%s", bailianExtra))
				// if err != nil {
				// 	return err
				// }
				return nil
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Infof("ragsync 版本 %s", app.Version)
		log.Info("使用 'ragsync help' 查看可用命令")
		log.Info("使用 'ragsync upload' 上传文件到百炼 RAG")
		log.Infof("aliyun ak: %v", utils.ShrinkString(c.GlobalString("access-key"), 5))
		log.Infof("aliyun sk: %v", utils.ShrinkString(c.GlobalString("secret"), 5))

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(utils.Errorf("运行 ragsync 出错: %v", err))
		os.Exit(1)
	}
}
