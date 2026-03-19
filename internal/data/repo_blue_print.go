package data

import "mockgitea/internal/models"

func BuildRepoBluePrint() []models.RepoBlueprint {
	blueprints := []models.RepoBlueprint{
		{OwnerLogin: "zhangsan", Name: "backend-api", Description: "核心后端 API 服务，承载认证、仓库同步和聚合查询。", Domain: "backend"},
		{OwnerLogin: "zhangsan", Name: "admin-console", Description: "运营后台，负责数据管理、报表和权限配置。", Domain: "frontend"},
		{OwnerLogin: "zhangsan", Name: "design-system", Description: "React 组件库与设计令牌，供多个前端项目复用。", Domain: "frontend"},
		{OwnerLogin: "lisi", Name: "mobile-app", Description: "面向业务团队的跨端移动应用。", Domain: "mobile"},
		{OwnerLogin: "lisi", Name: "flutter-shop", Description: "Flutter 电商客户端，覆盖商品、订单和会员中心。", Domain: "mobile"},
		{OwnerLogin: "lisi", Name: "miniprogram-portal", Description: "小程序门户，服务活动页与轻量业务流。", Domain: "mobile"},
		{OwnerLogin: "wangwu", Name: "data-pipeline", Description: "离线数据清洗与指标回流任务。", Domain: "data"},
		{OwnerLogin: "wangwu", Name: "analytics-service", Description: "指标查询和实时分析 API。", Domain: "backend"},
		{OwnerLogin: "wangwu", Name: "report-center", Description: "周报、日报与趋势分析中心。", Domain: "data"},
		{OwnerLogin: "zhaoliu", Name: "devops-platform", Description: "DevOps 门户，整合 CI、部署和环境管理。", Domain: "infra"},
		{OwnerLogin: "zhaoliu", Name: "terraform-modules", Description: "Terraform 模块仓库，统一管理基础设施模板。", Domain: "infra"},
		{OwnerLogin: "zhaoliu", Name: "k8s-manifests", Description: "Kubernetes 发布配置与环境差异化清单。", Domain: "infra"},
		{OwnerLogin: "chenxi", Name: "web-portal", Description: "客户门户 Web 站点。", Domain: "frontend"},
		{OwnerLogin: "chenxi", Name: "customer-h5", Description: "活动 H5 和营销落地页集合。", Domain: "frontend"},
		{OwnerLogin: "chenxi", Name: "docs-site", Description: "产品文档与接口说明站点。", Domain: "tooling"},
		{OwnerLogin: "yangfan", Name: "payment-gateway", Description: "支付聚合网关，对接多种支付通道。", Domain: "backend"},
		{OwnerLogin: "yangfan", Name: "order-service", Description: "订单域服务，负责履约状态流转。", Domain: "backend"},
		{OwnerLogin: "yangfan", Name: "inventory-service", Description: "库存域服务，处理预占、扣减和回补。", Domain: "backend"},
		{OwnerLogin: "zhoumo", Name: "ios-client", Description: "iOS 原生客户端。", Domain: "mobile"},
		{OwnerLogin: "zhoumo", Name: "android-client", Description: "Android 原生客户端。", Domain: "mobile"},
		{OwnerLogin: "zhoumo", Name: "react-native-kit", Description: "React Native 通用基础能力层。", Domain: "mobile"},
		{OwnerLogin: "sunqian", Name: "auth-service", Description: "统一认证授权中心。", Domain: "backend"},
		{OwnerLogin: "sunqian", Name: "openapi-sdk", Description: "OpenAPI SDK 生成与封装仓库。", Domain: "tooling"},
		{OwnerLogin: "sunqian", Name: "mock-data-factory", Description: "测试数据工厂，生成账号、订单和仓库数据。", Domain: "tooling"},
		{OwnerLogin: "wenyu", Name: "ci-templates", Description: "通用 CI 模板与流水线片段。", Domain: "tooling"},
		{OwnerLogin: "wenyu", Name: "monorepo-tools", Description: "Monorepo 工具链、脚本与规范。", Domain: "tooling"},
		{OwnerLogin: "wenyu", Name: "codegen", Description: "服务端和前端代码生成器。", Domain: "tooling"},
		{OwnerLogin: "linran", Name: "ai-assistant", Description: "内部智能助手与问答服务。", Domain: "ai"},
		{OwnerLogin: "linran", Name: "observability-stack", Description: "日志、指标、链路追踪一体化观测平台。", Domain: "infra"},
		{OwnerLogin: "linran", Name: "edge-proxy", Description: "边缘代理与灰度流量转发组件。", Domain: "infra"},
	}

	return blueprints
}