// import { UsersAPI } from '../request/users';
// import { HttpMgr } from '../request/http';
// import { DocumentAPI } from '../request/document';
// import { TEST_API_URL, TEST_TOKEN, TEST_UNAUTHORIZED } from './config';
//
// // 创建HTTP管理器实例
// const httpManager = new HttpMgr(TEST_API_URL, TEST_UNAUTHORIZED, TEST_TOKEN);
//
// // 创建UsersAPI实例
// const usersAPI = new UsersAPI(httpManager);
// const documentAPI = new DocumentAPI(httpManager);
//
// // 测试用户KV存储
// async function testGetKVStorage() {
//   try {
//     // 获取用户KV存储
//     const getResponse = await usersAPI.getKVStorage({ key: 'Preferences' });
//     console.log('获取用户KV存储:', getResponse);
//
//     // 设置用户KV存储
//     const setResponse = await usersAPI.setKVStorage({
//       key: 'Preferences',
//       value: '{"theme":"dark"}',
//     });
//     console.log('设置用户KV存储:', setResponse);
//   } catch (error) {
//     console.error('测试用户KV存储失败:', error);
//   }
// }
//
// // 测试用户文档访问记录
// // async function testUserDocumentAccessRecords() {
// //   try {
// //     // 获取用户文档访问记录列表
// //     const listResponse = await documentAPI.getUserDocumentAccessRecordsList();
// //     console.log('获取用户文档访问记录列表:', listResponse);
//
// //     if (listResponse.data && listResponse.data.length > 0) {
// //       // 删除第一条访问记录
// //       const deleteResponse = await documentAPI.deleteUserDocumentAccessRecord({
// //         access_record_id: listResponse.data[0].document_access_record.id,
// //       });
// //       console.log('删除用户文档访问记录:', deleteResponse);
// //     }
// //   } catch (error) {
// //     console.error('测试用户文档访问记录失败:', error);
// //   }
// // }
//
// // 测试用户反馈
// async function testSubmitFeedback() {
//   try {
//     // 提交用户反馈
//     const response = await usersAPI.submitFeedback({
//       type: 1,
//       content: '测试反馈内容',
//       page_url: 'https://example.com',
//       // image_path_list: ['image1.jpg'],
//     });
//     console.log('提交用户反馈:', response);
//   } catch (error) {
//     console.error('测试用户反馈失败:', error);
//   }
// }
//
// // setfilename
// async function testSetFilename() {
//   try {
//     const response = await documentAPI.setFileName({
//       doc_id: '1',
//       name: 'test',
//     });
//     console.log('设置文件名:', response);
//   } catch (error) {
//     console.error('设置文件名失败:', error);
//   }
// }
//
// // copyfile
// async function testCopyfile() {
//   try {
//     const response = await documentAPI.copyFile({
//       doc_id: '1',
//     });
//     console.log('复制文件:', response);
//   } catch (error) {
//     console.error('复制文件失败:', error);
//   }
// }
//
//
// // 运行所有测试
// async function runAllTests() {
//   console.log('开始测试用户相关API...');
//
//   // getInfo
//   // setAvatar
//   // setNickname
//   await testGetKVStorage();
//   // setKVStorage
//   await testSubmitFeedback();
//   // refreshToken
//
//   console.log('测试完成');
// }
//
// // 执行测试
// runAllTests().catch(console.error);
//
// function testSetfavoriteStatus() {
//   throw new Error('Function not implemented.');
// }
