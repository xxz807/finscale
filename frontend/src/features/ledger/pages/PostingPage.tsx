import { useState } from 'react';
import { Form, Input, Button, Select, Card, Space, Typography, Row, Col, Divider } from 'antd';
import { MinusCircleOutlined, PlusOutlined, BankOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { ledgerApi, type PostTransactionReq } from '../api/ledgerApi';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

export const PostingPage = () => {
  const [form] = Form.useForm();

  // 生成随机 Reference ID (模拟)
  const generateRefId = () => `TX-${Date.now()}-${Math.floor(Math.random() * 1000)}`;

  // === 1. 核心逻辑: React Query Mutation ===
  const mutation = useMutation({
    mutationFn: ledgerApi.postTransaction,
    onSuccess: (data) => {
      // 成功后弹窗已经在 api client 里统一处理了，这里只需要重置表单
      // 你也可以在这里加一个 message.success("记账成功")
      form.resetFields();
      form.setFieldValue('reference_id', generateRefId());
    },
    // onError 也在 api client 里处理了，这里不需要写
  });

  const onFinish = (values: PostTransactionReq) => {
    // 触发提交
    mutation.mutate(values);
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <Card
        title={<Space><BankOutlined /><span>Core Banking General Ledger</span></Space>}
        extra={<Text type="secondary">System: FinScale v1.0</Text>}
        className="shadow-lg rounded-xl"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            reference_id: generateRefId(),
            tx_type: 'TRANSFER',
            // 默认两条分录：一借一贷
            postings: [
              { direction: 'D', amount: '' },
              { direction: 'C', amount: '' }
            ]
          }}
        >
          {/* === Header Section: 交易主信息 === */}
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="reference_id"
                label="Reference ID (Idempotency Key)"
                rules={[{ required: true }]}
              >
                <Input prefix="#" disabled className="bg-gray-50" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="tx_type"
                label="Transaction Type"
                rules={[{ required: true }]}
              >
                <Select>
                  <Option value="TRANSFER">Transfer (转账)</Option>
                  <Option value="DEPOSIT">Deposit (存款)</Option>
                  <Option value="WITHDRAW">Withdraw (提现)</Option>
                  <Option value="FEE">Fee (手续费)</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="Description">
            <TextArea rows={2} placeholder="e.g. Monthly salary payment" />
          </Form.Item>

          <Divider orientation="left">Journal Entries (会计分录)</Divider>

          {/* === Lines Section: 动态分录列表 === */}
          <Form.List name="postings">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Row key={key} gutter={12} align="middle" className="mb-2">
                    <Col span={8}>
                      <Form.Item
                        {...restField}
                        name={[name, 'account_code']}
                        rules={[{ required: true, message: 'Missing account' }]}
                        className="mb-0"
                      >
                        <Input placeholder="Account Code (e.g. 1001)" />
                      </Form.Item>
                    </Col>

                    <Col span={6}>
                      <Form.Item
                        {...restField}
                        name={[name, 'direction']}
                        rules={[{ required: true }]}
                        className="mb-0"
                      >
                        <Select>
                          <Option value="D">Debit (借)</Option>
                          <Option value="C">Credit (贷)</Option>
                        </Select>
                      </Form.Item>
                    </Col>

                    <Col span={8}>
                      <Form.Item
                        {...restField}
                        name={[name, 'amount']}
                        rules={[{ required: true, message: 'Missing amount' }]}
                        className="mb-0"
                      >
                        <Input
                          prefix="¥"
                          placeholder="0.00"
                          // 只能输入数字和小数点
                          onChange={(e) => {
                            const { value } = e.target;
                            const reg = /^-?\d*(\.\d*)?$/;
                            if (!reg.test(value)) {
                              // 实际项目中建议用 InputNumber 组件，这里为了演示 String 传参逻辑
                            }
                          }}
                        />
                      </Form.Item>
                    </Col>

                    <Col span={2}>
                      <MinusCircleOutlined
                        className="text-red-500 cursor-pointer hover:text-red-700"
                        onClick={() => remove(name)}
                      />
                    </Col>
                  </Row>
                ))}

                <Form.Item className="mt-4">
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    Add Entry Line
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Divider />

          {/* === Submit Button === */}
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              block
              loading={mutation.isPending} // 自动处理 Loading 状态
              className="bg-blue-600 hover:bg-blue-500"
            >
              Post Transaction
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};