import { UsersAPI } from '../request/users';
import { HttpMgr } from '../request/http';
import { TEST_API_URL, TEST_TOKEN, TEST_UNAUTHORIZED } from './config';

// 创建HTTP管理器实例
const httpManager = new HttpMgr(TEST_API_URL, TEST_UNAUTHORIZED, TEST_TOKEN);

// 创建UsersAPI实例
const usersAPI = new UsersAPI(httpManager);

// 测试用户KV存储
async function testUserKVStorage() {
  try {
    // 获取用户KV存储
    const getResponse = await usersAPI.getUserKVStorageAPI({ key: 'Preferences' });
    console.log('获取用户KV存储:', getResponse);

    // 设置用户KV存储
    const setResponse = await usersAPI.setUserKVStorageAPI({
      key: 'Preferences',
      value: '{"theme":"dark"}',
    });
    console.log('设置用户KV存储:', setResponse);
  } catch (error) {
    console.error('测试用户KV存储失败:', error);
  }
}

// 测试用户文档访问记录
async function testUserDocumentAccessRecords() {
  try {
    // 获取用户文档访问记录列表
    const listResponse = await usersAPI.getUserDocumentAccessRecordsListAPI();
    console.log('获取用户文档访问记录列表:', listResponse);

    if (listResponse.data && listResponse.data.length > 0) {
      // 删除第一条访问记录
      const deleteResponse = await usersAPI.deleteUserDocumentAccessRecordAPI({
        access_record_id: listResponse.data[0].document_access_record.id,
      });
      console.log('删除用户文档访问记录:', deleteResponse);
    }
  } catch (error) {
    console.error('测试用户文档访问记录失败:', error);
  }
}

// 测试用户反馈
async function testUserFeedback() {
  try {
    // 提交用户反馈
    const response = await usersAPI.submitFeedbackAPI({
      type: 1,
      content: '测试反馈内容',
      page_url: 'https://example.com',
      // image_path_list: ['image1.jpg'],
    });
    console.log('提交用户反馈:', response);
  } catch (error) {
    console.error('测试用户反馈失败:', error);
  }
}

// 运行所有测试
async function runAllTests() {
  console.log('开始测试用户相关API...');
  
  await testUserKVStorage();
  await testUserDocumentAccessRecords();
  await testUserFeedback();
  
  console.log('测试完成');
}

// 执行测试
runAllTests().catch(console.error); 