import { Table, Tag, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { ledgerApi, type Account } from '../api/ledgerApi';

const { Text } = Typography;

export const AccountList = () => {
  // 使用 React Query 获取数据
  // queryKey: ['accounts'] 是缓存的唯一标识
  const { data, isLoading } = useQuery({
    queryKey: ['accounts'],
    queryFn: ledgerApi.getAccounts,
  });

  const columns = [
    {
      title: 'Code',
      dataIndex: 'AccountCode',
      key: 'AccountCode',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: 'Name',
      dataIndex: 'Name',
      key: 'Name',
    },
    {
      title: 'Type',
      dataIndex: 'Type',
      key: 'Type',
      render: (type: number) => {
        const map = { 1: 'Asset', 2: 'Liability', 3: 'Equity', 4: 'Income', 5: 'Expense' };
        const colors = { 1: 'blue', 2: 'orange', 3: 'purple', 4: 'green', 5: 'red' };
        // @ts-ignore
        return <Tag color={colors[type]}>{map[type] || 'Unknown'}</Tag>;
      }
    },
    {
      title: 'Balance',
      dataIndex: 'Balance',
      key: 'Balance',
      align: 'right' as const,
      render: (balance: string, record: Account) => {
        // 简单的金额格式化
        const val = parseFloat(balance);
        const color = val < 0 ? 'text-red-500' : 'text-gray-800';
        return (
          <span className={`font-mono font-bold ${color}`}>
            {new Intl.NumberFormat('zh-CN', { minimumFractionDigits: 2 }).format(val)} {record.Currency}
          </span>
        );
      },
    },
  ];

  return (
    <Table 
      dataSource={data} 
      columns={columns} 
      rowKey="ID"
      loading={isLoading}
      pagination={false}
      className="mt-6 border rounded-lg overflow-hidden"
      size="small"
    />
  );
};