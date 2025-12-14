import { App as AntdApp, ConfigProvider, Button, Card } from 'antd';
import { StyleProvider } from '@ant-design/cssinjs';
import 'antd/dist/reset.css'; // AntD 6 依然建议引入 reset，虽然它有 cssinjs

function App() {
  return (
    // 1. StyleProvider: 提高 AntD 样式的优先级，防止被 Tailwind 覆盖
    <StyleProvider hashPriority="high">
      {/* 2. ConfigProvider: 配置全局主题 */}
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#1677ff', // 支付宝蓝
          },
        }}
      >
        {/* 3. AntdApp: 提供全局上下文 (Message, Modal 等) */}
        <AntdApp>
          <div className="min-h-screen flex items-center justify-center bg-gray-100">
            <Card className="w-96 shadow-xl rounded-xl border-none">
              <h1 className="text-2xl font-bold text-gray-800 mb-4">FinScale Console</h1>
              <p className="text-gray-500 mb-6">
                System Status: <span className="text-green-500 font-semibold">Online</span>
              </p>
              
              <div className="flex gap-4">
                {/* 测试 AntD 组件与 Tailwind 类名混用 */}
                <Button type="primary" className="bg-blue-600">
                  AntD Button
                </Button>
                <button className="px-4 py-1 bg-white border border-gray-300 rounded hover:bg-gray-50">
                  Tailwind Button
                </button>
              </div>
            </Card>
          </div>
        </AntdApp>
      </ConfigProvider>
    </StyleProvider>
  );
}

export default App;