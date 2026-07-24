const repoRoot = 'https://github.com/alibaba/loongsuite-go';
export default {
    lang: 'en-US',
    title: ' ',
    description: 'It provides an automatic solution for Golang applications that want to leverage OpenTelemetry to enable effective observability. No code changes are required in the target application, the instrumentation is done at compile time. Simply adding `otel` prefix to `go build` to get started ',
    ignoreDeadLinks: true,
    base: '/loongsuite-go/',
    locales: {
        root: {
          label: 'English',
          lang: 'en-US',
        },
        zh: {
          label: '简体中文',
          lang: 'zh-CN',
          link: '/zh/',
        },
    },
    themeConfig: {
        logo: '/loongsuite.png',
        nav: [
            {
                text: 'Download',
                items: [
                    { text: 'Linux AMD64', link: `${repoRoot}/releases/latest/download/otel-linux-amd64` },
                    { text: 'Linux ARM64', link: `${repoRoot}/releases/latest/download/otel-linux-arm64` },
                    { text: 'MacOS AMD64', link: `${repoRoot}/releases/latest/download/otel-darwin-amd64` },
                    { text: 'MacOS ARM64', link: `${repoRoot}/releases/latest/download/otel-darwin-arm64` },
                    { text: 'Windows AMD64', link: `${repoRoot}/releases/latest/download/otel-windows-amd64.exe` },
                ]
            },
            {
                text: 'Other Agents',
                items: [
                    { text: 'Go', link: 'https://github.com/alibaba/loongsuite-go' },
                    { text: 'Java', link: 'https://github.com/alibaba/loongsuite-java-agent' },
                    { text: 'Python', link: 'https://github.com/alibaba/loongsuite-python' },
                ]
            }
        ],
        socialLinks: [
            { icon: 'github', link: repoRoot },
            { icon: 'alibabacloud', link: 'https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications' }
        ],
        editLink: {
            pattern: `${repoRoot}/edit/main/docs/:path`
        },
        sidebar: {
            '/': [
                {
                  text: '🌟 User Guide',
                  items: [
                    { text: 'Overview', link: '/index' },
                    { text: 'CLI Reference', link: '/user/config' },
                    { text: 'Compilation Time', link: '/user/compilation-time' },
                    { text: 'Experimental', link: '/user/experimental-feature' },
                    { text: 'Compatibility', link: '/user/compatibility' },
                    { text: 'Manual Instrumentation', link: '/user/manual-instrument' },
                    { text: 'Context Propagation', link: '/user/context-propagation' },
                    { text: 'OpenTelemetry Settings', link: '/user/sdk-config' },
                    { text: 'Supported Libraries', link: '/user/supported-libraries' },
                  ]
                },
                {
                    text: '🔧 Developer Guide',
                    items: [
                        { text: 'Overview', link: '/dev/overview' },
                        { text: 'Register Hook Rule', link: '/dev/register' },
                        { text: 'Write the Hook Code', link: '/dev/hook' },
                        { text: 'Test the Hook Code', link: '/dev/test' },
                        { text: 'Hook Rule Types', link: '/dev/rule_def' },
                        { text: 'HTTP Capture Proposal', link: '/dev/http-capture' },
                    ]
                  },
                {
                    text: '🤠 Hacking Guide',
                    items: [
                      { text: 'Overview', link: '/hacking/overview' },
                      { text: 'Preprocess Phase', link: '/hacking/preprocess' },
                      { text: 'Instrument Phase', link: '/hacking/instrument' },
                      { text: 'AST Optimization', link: '/hacking/optimize' },
                      { text: 'Debugging', link: '/hacking/debug' },
                      { text: 'Tool Internal Slides', link: 'https://github.com/alibaba/loongsuite-go/blob/main/docs/otel-alibaba.pdf' },
                    ]
                },
                {
                    text: '🌐 Community',
                    items: [
                        { text: 'DingTalk', link: 'https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11' },
                    ]
                },
            ],
            '/zh/': [
                {
                  text: '🌟 用户指南',
                  items: [
                    { text: '概述', link: '/zh/index' },
                    { text: '命令行参考', link: '/zh/user/config' },
                    { text: '编译时间', link: '/zh/user/compilation-time' },
                    { text: '实验性功能', link: '/zh/user/experimental-feature' },
                    { text: '兼容性', link: '/zh/user/compatibility' },
                    { text: '手动埋点', link: '/zh/user/manual-instrument' },
                    { text: '上下文传播', link: '/zh/user/context-propagation' },
                    { text: 'OpenTelemetry 配置', link: '/zh/user/sdk-config' },
                    { text: '支持的库', link: '/zh/user/supported-libraries' },
                  ]
                },
                {
                    text: '🔧开发者指南',
                    items: [
                        { text: '概述', link: '/zh/dev/overview' },
                        { text: '注册Hook规则', link: '/zh/dev/register' },
                        { text: '编写Hook代码', link: '/zh/dev/hook' },
                        { text: '测试Hook代码', link: '/zh/dev/test' },
                        { text: 'Hook规则类型', link: '/zh/dev/rule_def' },
                        { text: 'HTTP采集方案', link: '/zh/dev/http-capture' },
                    ]
                  },
                {
                    text: '🤠 黑客指南',
                    items: [
                      { text: '概述', link: '/zh/hacking/overview' },
                      { text: '预处理阶段', link: '/zh/hacking/preprocess' },
                      { text: '埋点阶段', link: '/zh/hacking/instrument' },
                      { text: 'AST优化', link: '/zh/hacking/optimize' },
                      { text: '调试', link: '/zh/hacking/debug' },
                      { text: '工具内幕幻灯片', link: 'https://github.com/alibaba/loongsuite-go/blob/main/docs/otel-alibaba.pdf' },
                    ]
                },
                {
                    text: '🌐 社区',
                    items: [
                        { text: '钉钉', link: 'https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11' },
                    ]
                },
            ]
        }
    }
}
