package component

// MessageCode is a stable, frontend-facing identifier for i18n messages.
// Backend responses expose messageCode and params only; clients translate them locally.
type MessageCode string

// MessageParams contains dynamic values used by frontend translations.
type MessageParams map[string]any

// MessageError carries a stable code plus fallback text through service helpers.
type MessageError struct {
	Code     MessageCode
	Fallback string
	Params   MessageParams
}

func (err MessageError) Error() string {
	return err.Fallback
}

func NewMessageError(code MessageCode, fallback string, params MessageParams) error {
	return MessageError{
		Code:     code,
		Fallback: fallback,
		Params:   params,
	}
}

const (
	MessageRequestInvalidFormat MessageCode = "common.request.invalidFormat" // 请求体或参数格式无法解析。
	MessageRequestInvalidParams MessageCode = "common.request.invalidParams" // 请求参数未通过业务校验。
	MessageRequestParseFailed   MessageCode = "common.request.parseFailed"   // 参数绑定失败，params.error 可带原始错误。
	MessageOperationSuccess     MessageCode = "common.operation.success"     // 通用操作成功。
	MessageOperationFailed      MessageCode = "common.operation.failed"      // 通用操作失败。
	MessagePageNotFound         MessageCode = "page.notFound"                // 页面不存在。
	MessageRouteNotFound        MessageCode = "route.notFound"               // 路由未定义。
	MessageUserFetchFailed      MessageCode = "user.fetchFailed"             // 当前用户信息读取失败。
	MessageUserNotFound         MessageCode = "user.notFound"                // 用户不存在。
	MessageUserUpdateFailed     MessageCode = "user.updateFailed"            // 用户信息保存失败。
	MessageUserUpdateSuccess    MessageCode = "user.updateSuccess"           // 用户信息保存成功。
)

const (
	MessageAuthRequired                  MessageCode = "auth.required"                   // 需要登录后才能继续操作。
	MessageAuthSignupDisabled            MessageCode = "auth.signupDisabled"             // 当前站点关闭了注册。
	MessageAuthEmailDomainInvalid        MessageCode = "auth.emailDomain.invalid"        // 邮箱格式不正确或无法提取域名。
	MessageAuthEmailDomainNotAllowed     MessageCode = "auth.emailDomain.notAllowed"     // 邮箱域名不在注册白名单。
	MessageAuthUsernameInvalid           MessageCode = "auth.username.invalid"           // 用户名格式不符合规则。
	MessageAuthUsernameExists            MessageCode = "auth.username.exists"            // 用户名已存在。
	MessageAuthEmailExists               MessageCode = "auth.email.exists"               // 邮箱已被使用。
	MessageAuthPasswordTooShort          MessageCode = "auth.password.tooShort"          // 密码过短，params.minLength 表示最小长度。
	MessageAuthPasswordTooLong           MessageCode = "auth.password.tooLong"           // 密码过长。
	MessageAuthPasswordNeedsLetterNumber MessageCode = "auth.password.needsLetterNumber" // 密码必须包含字母和数字。
	MessageAuthCaptchaInvalid            MessageCode = "auth.captcha.invalid"            // 验证码错误或已过期。
	MessageAuthRegisterFailed            MessageCode = "auth.register.failed"            // 注册失败。
	MessageAuthRegisterRetryLogin        MessageCode = "auth.register.retryLogin"        // 注册成功但自动登录失败，建议手动登录。
	MessageAuthRegisterEmailVerify       MessageCode = "auth.register.emailVerify"       // 注册成功，需要邮箱验证。
	MessageAuthLoginSuccess              MessageCode = "auth.login.success"              // 登录成功。
	MessageAuthLoginInvalidRequest       MessageCode = "auth.login.invalidRequest"       // 登录请求无效，通常需要刷新页面重试。
	MessageAuthPasswordInvalidFormat     MessageCode = "auth.password.invalidFormat"     // 登录密码格式不正确。
	MessageAuthInvalidCredentials        MessageCode = "auth.credentials.invalid"        // 用户名、邮箱或密码错误。
	MessageAuthAccountFrozen             MessageCode = "auth.account.frozen"             // 账号被冻结。
	MessageAuthEmailUnverified           MessageCode = "auth.email.unverified"           // 邮箱未验证。
	MessageAuthLoginFailed               MessageCode = "auth.login.failed"               // 登录异常。
	MessageAuthOldPasswordInvalid        MessageCode = "auth.password.oldInvalid"        // 原密码错误。
	MessageAuthPasswordUpdateFailed      MessageCode = "auth.password.updateFailed"      // 修改密码失败。
	MessageAuthPasswordUpdateSuccess     MessageCode = "auth.password.updateSuccess"     // 修改密码成功。
	MessageAuthResetMailQueued           MessageCode = "auth.passwordReset.mailQueued"   // 如邮箱存在，将收到密码重置邮件。
	MessageAuthResetTokenCreateFailed    MessageCode = "auth.passwordReset.tokenFailed"  // 生成重置令牌失败。
	MessageAuthResetMailSendFailed       MessageCode = "auth.passwordReset.mailFailed"   // 发送重置邮件失败。
	MessageAuthResetTokenInvalid         MessageCode = "auth.passwordReset.tokenInvalid" // 重置链接过期或无效。
	MessageAuthResetFailed               MessageCode = "auth.passwordReset.failed"       // 重置密码失败。
	MessageAuthResetSuccess              MessageCode = "auth.passwordReset.success"      // 重置密码成功。
	MessageAuthActivationResendSuccess   MessageCode = "auth.activation.resendSuccess"   // 验证邮件已重新发送。
	MessageAuthActivationAlreadyVerified MessageCode = "auth.activation.alreadyVerified" // 当前账号已完成邮箱验证。
	MessageAuthActivationDisabled        MessageCode = "auth.activation.disabled"        // 当前站点未启用邮箱验证。
	MessageAuthActivationResendCooldown  MessageCode = "auth.activation.resendCooldown"  // 验证邮件发送过于频繁，params.retryAfterSeconds。
	MessageAuthActivationResendDaily     MessageCode = "auth.activation.resendDaily"     // 验证邮件达到当天重发上限，params.limit。
	MessageAuthActivationResendFailed    MessageCode = "auth.activation.resendFailed"    // 验证邮件重新发送失败。
)

const (
	MessagePermissionResolveFailed MessageCode = "permission.resolveFailed" // 权限信息读取失败。
	MessagePermissionDenied        MessageCode = "permission.denied"        // 当前用户没有执行该操作的权限。
	MessagePermissionUserFrozen    MessageCode = "permission.userFrozen"    // 用户已被冻结，params.action 表示操作名称。
	MessagePermissionEmailRequired MessageCode = "permission.emailRequired" // 需要先完成邮箱验证，params.action 表示操作名称。
)

const (
	MessageUploadAttachmentDisabled MessageCode = "upload.attachment.disabled"   // 站点关闭了附件上传。
	MessageUploadCooldown           MessageCode = "upload.cooldown"              // 新用户上传冷却中，params.minutes/availableAt。
	MessageUploadDailyLimit         MessageCode = "upload.dailyLimit"            // 达到每日上传限制，params.count。
	MessageUploadDailyLimitAvatar   MessageCode = "upload.dailyLimit.avatar"     // 头像上传将超过每日限制，params.count/fileCount。
	MessageUploadFileMissing        MessageCode = "upload.file.missing"          // 未获取到上传文件。
	MessageUploadFilenameRequired   MessageCode = "upload.filename.required"     // 文件名为空。
	MessageUploadFileTooLarge       MessageCode = "upload.file.tooLarge"         // 文件超过大小限制，params.maxSizeKb。
	MessageUploadUnsupportedExt     MessageCode = "upload.extension.unsupported" // 文件扩展名不允许，params.extensions。
	MessageUploadUnsupportedImage   MessageCode = "upload.image.unsupported"     // 图片格式不支持。
	MessageUploadReadFailed         MessageCode = "upload.readFailed"            // 文件读取失败。
	MessageUploadOpenFailed         MessageCode = "upload.openFailed"            // 文件打开失败。
	MessageUploadInvalidImage       MessageCode = "upload.image.invalidContent"  // 文件内容不是有效图片。
	MessageUploadContentReadFailed  MessageCode = "upload.contentReadFailed"     // 文件内容读取失败。
	MessageUploadSaveFailed         MessageCode = "upload.saveFailed"            // 文件保存失败，params.error 可带原始错误。
	MessageUploadSuccess            MessageCode = "upload.success"               // 上传成功。
)

const (
	MessageArticleNotFound          MessageCode = "article.notFound"          // 文章不存在。
	MessageArticleOwnerMismatch     MessageCode = "article.ownerMismatch"     // 不能修改或删除他人的文章。
	MessageArticleOperationDenied   MessageCode = "article.operationDenied"   // 当前文章不可操作。
	MessageArticleSaveFailed        MessageCode = "article.saveFailed"        // 文章保存失败。
	MessageArticleDailyLimit        MessageCode = "article.dailyLimit"        // 当天发布过多。
	MessageArticleTitleTooShort     MessageCode = "article.title.tooShort"    // 标题过短，params.minLength。
	MessageArticleTitleTooLong      MessageCode = "article.title.tooLong"     // 标题过长，params.maxLength。
	MessageArticleContentTooShort   MessageCode = "article.content.tooShort"  // 正文过短，params.minLength。
	MessageArticleContentTooLong    MessageCode = "article.content.tooLong"   // 正文过长，params.maxLength。
	MessageArticlePostCooldown      MessageCode = "article.post.cooldown"     // 新用户发帖冷却中，params.minutes/availableAt。
	MessageCommentContentTooShort   MessageCode = "comment.content.tooShort"  // 评论过短，params.minLength。
	MessageCommentContentTooLong    MessageCode = "comment.content.tooLong"   // 评论过长，params.maxLength。
	MessageCommentPostCooldown      MessageCode = "comment.post.cooldown"     // 新用户评论冷却中，params.minutes/availableAt。
	MessageCommentReplyTargetMissed MessageCode = "comment.replyTargetMissed" // 被回复的评论不存在。
	MessageCommentCreateFailed      MessageCode = "comment.createFailed"      // 评论创建失败，params.error 可带原始错误。
	MessageReplyNotFound            MessageCode = "reply.notFound"            // 回复不存在。
	MessageReplyUpdateFailed        MessageCode = "reply.updateFailed"        // 回复更新失败，params.error 可带原始错误。
	MessageReportNotFound           MessageCode = "report.notFound"           // 举报不存在。
	MessageReportTargetInvalid      MessageCode = "report.targetInvalid"      // 举报对象无效。
	MessageReportOwnContent         MessageCode = "report.ownContent"         // 不能举报自己的内容。
	MessageReportDuplicate          MessageCode = "report.duplicate"          // 已举报，等待处理。
	MessageReportCreateFailed       MessageCode = "report.createFailed"       // 举报提交失败。
)

const (
	MessageNotificationMarkReadFailed  MessageCode = "notification.markRead.failed"     // 标记单条通知已读失败。
	MessageNotificationMarkReadSuccess MessageCode = "notification.markRead.success"    // 标记单条通知已读成功。
	MessageNotificationMarkAllFailed   MessageCode = "notification.markAllRead.failed"  // 标记全部通知已读失败。
	MessageNotificationMarkAllSuccess  MessageCode = "notification.markAllRead.success" // 标记全部通知已读成功。
	MessageOAuthUnbindFailed           MessageCode = "oauth.unbind.failed"              // 解绑第三方账号失败，params.error 可带原始错误。
	MessageOAuthUnbindSuccess          MessageCode = "oauth.unbind.success"             // 解绑第三方账号成功。
	MessageOAuthCallbackFailed         MessageCode = "oauth.callback.failed"            // OAuth 认证回调失败。
	MessageOAuthProcessFailed          MessageCode = "oauth.process.failed"             // OAuth 登录处理失败。
	MessageOAuthAccountFrozen          MessageCode = "oauth.account.frozen"             // OAuth 登录账号被冻结。
	MessageOAuthActivationUpdateFailed MessageCode = "oauth.activation.updateFailed"    // OAuth 用户激活状态更新失败。
	MessageOAuthTokenFailed            MessageCode = "oauth.token.failed"               // OAuth 登录 token 生成失败。
	MessageChatSendFailed              MessageCode = "chat.send.failed"                 // 私信发送失败，params.error 可带原始错误。
	MessageChatGetMessagesFailed       MessageCode = "chat.messages.failed"             // 获取私信列表失败。
	MessageChatMarkReadFailed          MessageCode = "chat.markRead.failed"             // 标记私信已读失败。
)

const (
	MessageAdminStatsFetchFailed       MessageCode = "admin.stats.fetchFailed"         // 管理后台统计数据读取失败。
	MessageAdminBadgeNameRequired      MessageCode = "admin.badge.nameRequired"        // 徽章名称不能为空。
	MessageAdminBadgeTypeInvalid       MessageCode = "admin.badge.typeInvalid"         // 徽章类型不合法。
	MessageAdminBadgeCodeRequired      MessageCode = "admin.badge.codeRequired"        // 徽章编码不能为空。
	MessageAdminBadgeGrantModeInvalid  MessageCode = "admin.badge.grantModeInvalid"    // 徽章授予方式不合法。
	MessageAdminBadgeSystemNotFound    MessageCode = "admin.badge.systemNotFound"      // 系统徽章不存在。
	MessageAdminBadgeSaveFailed        MessageCode = "admin.badge.saveFailed"          // 保存徽章失败。
	MessageAdminBadgeSystemDeleteBlock MessageCode = "admin.badge.systemDeleteBlocked" // 系统默认徽章不可删除。
	MessageAdminBadgeDeleteFailed      MessageCode = "admin.badge.deleteFailed"        // 删除徽章失败。
	MessageAdminTargetUserFetchFailed  MessageCode = "admin.user.targetFetchFailed"    // 目标用户查询失败。
	MessageAdminCategoryRequired       MessageCode = "admin.category.nameRequired"     // 分类名称不能为空。
	MessageAdminCategoryNotFound       MessageCode = "admin.category.notFound"         // 分类不存在。
	MessageAdminCategoryDataNotFound   MessageCode = "admin.category.dataNotFound"     // 分类数据不存在。
	MessageAdminCategoryKeepOne        MessageCode = "admin.category.keepOne"          // 至少保留一个分类。
	MessageAdminCategoryHasArticles    MessageCode = "admin.category.hasArticles"      // 分类下存在有效文章。
	MessageAdminModeratorUserRequired  MessageCode = "admin.moderator.userRequired"    // 版主用户不能为空。
	MessageAdminModeratorUserNotFound  MessageCode = "admin.moderator.userNotFound"    // 版主用户不存在。
	MessageAdminModeratorNotFound      MessageCode = "admin.moderator.notFound"        // 版主记录不存在。
	MessageAdminCategorySelectRequired MessageCode = "admin.article.categoryRequired"  // 文章至少需要一个分类。
	MessageAdminCategorySelectTooMany  MessageCode = "admin.article.categoryTooMany"   // 文章最多选择三个分类。
	MessageAdminArticleDeleteFailed    MessageCode = "admin.article.deleteFailed"      // 删除文章失败。
	MessageAdminRoleNotFound           MessageCode = "admin.role.notFound"             // 角色不存在。
	MessageAdminTestEmailRequired      MessageCode = "admin.mail.testEmailRequired"    // 测试邮箱不能为空。
	MessageAdminTestEmailFailed        MessageCode = "admin.mail.testFailed"           // 邮件配置测试失败，params.error 可带原始错误。
	MessageAdminTestEmailSuccess       MessageCode = "admin.mail.testSuccess"          // 邮件配置测试成功，params.email 表示测试邮箱。
)
