package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "user:manage",
		Short: "交互式用户管理工具",
		Run:   runUserManagement,
	}
	appendCommand(cmd)
}

func runUserManagement(cmd *cobra.Command, _ []string) {
	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动信号监听协程
	go func() {
		<-sigChan
		fmt.Println("\n\n收到退出信号，正在优雅关闭...")
		cancel()
	}()

	fmt.Println("=== 用户管理工具 ===")
	fmt.Println("输入 'help' 查看帮助信息")
	fmt.Println("输入 'quit' 或 'exit' 退出程序")
	fmt.Println("使用 Ctrl+C 也可以安全退出")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("程序已安全退出")
			return
		default:
			fmt.Print("> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("读取输入错误: %v\n", err)
				continue
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			if handleCommand(input, ctx) {
				fmt.Println("再见！")
				return
			}
		}
	}
}

func handleCommand(input string, ctx context.Context) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "help", "h":
		showHelp()
	case "quit", "exit", "q":
		return true
	case "list", "ls":
		listUsers()
	case "info":
		if len(parts) < 2 {
			fmt.Println("请提供用户ID: info <userId>")
			return false
		}
		userID, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			fmt.Printf("无效的用户ID: %s\n", parts[1])
			return false
		}
		showUserInfo(userID)
	case "password", "pwd":
		changePasswordInteractive()
	case "admin":
		setAdminInteractive()
	case "clear":
		fmt.Print("\033[2J\033[H") // 清屏
	default:
		fmt.Printf("未知命令: %s\n", command)
		fmt.Println("输入 'help' 查看帮助信息")
	}

	return false
}

func showHelp() {
	fmt.Println("\n=== 可用命令 ===")
	fmt.Println("help, h        - 显示帮助信息")
	fmt.Println("list, ls       - 列出所有用户")
	fmt.Println("info <userId>  - 显示用户详细信息")
	fmt.Println("password, pwd  - 修改用户密码")
	fmt.Println("admin          - 设置用户为管理员")
	fmt.Println("clear          - 清屏")
	fmt.Println("quit, exit, q  - 退出程序")
	fmt.Println()
}

func listUsers() {
	fmt.Println("\n=== 用户列表 ===")
	// 这里需要根据实际的用户模型来实现
	// 假设有一个获取所有用户的方法
	fmt.Println("功能待实现：列出所有用户")
	fmt.Println("提示：使用 'info <userId>' 查看特定用户信息")
	fmt.Println()
}

func showUserInfo(userID uint64) {
	userEntity, err := users.Get(userID)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}

	if userEntity.Id == 0 {
		fmt.Printf("用户ID %d 不存在\n", userID)
		return
	}

	fmt.Printf("\n=== 用户信息 ===\n")
	fmt.Printf("用户ID: %d\n", userEntity.Id)
	fmt.Printf("用户名: %s\n", userEntity.Username)
	fmt.Printf("邮箱: %s\n", userEntity.Email)
	fmt.Printf("角色ID: %d\n", userEntity.RoleId)
	fmt.Printf("状态: %d\n", userEntity.Status)
	fmt.Println()
}

func changePasswordInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入用户ID: ")
	userIDStr, _ := reader.ReadString('\n')
	userIDStr = strings.TrimSpace(userIDStr)

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的用户ID: %s\n", userIDStr)
		return
	}

	userEntity, err := users.Get(userID)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}

	if userEntity.Id == 0 {
		fmt.Printf("用户ID %d 不存在\n", userID)
		return
	}

	fmt.Printf("当前用户: %s (ID: %d)\n", userEntity.Username, userEntity.Id)
	fmt.Print("请输入新密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		fmt.Println("密码不能为空")
		return
	}

	fmt.Print("确认修改密码？(y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("操作已取消")
		return
	}

	userEntity.SetPassword(password)
	r := users.Save(&userEntity)
	if r != nil {
		fmt.Printf("密码修改失败: %v\n", r)
	} else {
		fmt.Println("密码修改成功！")
	}
}

func setAdminInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入用户ID: ")
	userIDStr, _ := reader.ReadString('\n')
	userIDStr = strings.TrimSpace(userIDStr)

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的用户ID: %s\n", userIDStr)
		return
	}

	userEntity, err := users.Get(userID)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}

	if userEntity.Id == 0 {
		fmt.Printf("用户ID %d 不存在\n", userID)
		return
	}

	fmt.Printf("当前用户: %s (ID: %d)\n", userEntity.Username, userEntity.Id)
	fmt.Print("确认设置为管理员？(y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("操作已取消")
		return
	}

	// 判断有没有管理员角色
	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		role.SaveOrCreateById(&roleEntity)
		fmt.Println("角色不存在，已创建管理员角色")
	}

	rp := rolePermissionRs.GetRsByRoleIdAndPermission(roleEntity.Id, permission.Admin.Id())
	if rp.Id == 0 {
		rp.RoleId = roleEntity.Id
		rp.PermissionId = permission.Admin.Id()
		rp.Effective = 1
		rolePermissionRs.SaveOrCreateById(&rp)
		fmt.Println("角色权限关系不存在，已创建角色权限关系")
	}

	userEntity.RoleId = roleEntity.Id
	err = users.Save(&userEntity)
	if err != nil {
		fmt.Printf("设置管理员失败: %v\n", err)
	} else {
		fmt.Println("用户已成功设置为管理员！")
	}
}
