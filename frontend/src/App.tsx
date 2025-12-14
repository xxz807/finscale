import { App as AntdApp, ConfigProvider } from 'antd';
import { StyleProvider } from '@ant-design/cssinjs';
import 'antd/dist/reset.css';
import { PostingPage } from './features/ledger/pages/PostingPage';

function App() {
  return (
    <StyleProvider hashPriority="high">
      <ConfigProvider
        theme={{
          token: { colorPrimary: '#1677ff' },
        }}
      >
        <AntdApp>
          <div className="min-h-screen bg-gray-100 py-10">
            {/* 直接渲染记账页 */}
            <PostingPage />
          </div>
        </AntdApp>
      </ConfigProvider>
    </StyleProvider>
  );
}

export default App;