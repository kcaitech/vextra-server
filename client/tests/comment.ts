/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

// import { CommentAPI } from '../request/comment'
// import { HttpMgr } from '../request/http'
// import { TEST_API_URL, TEST_TOKEN, TEST_UNAUTHORIZED } from './config';
//
// // 创建HTTP管理器实例
// const httpManager = new HttpMgr(TEST_API_URL, TEST_UNAUTHORIZED, TEST_TOKEN);
//
// // 创建CommentAPI实例
// const commentAPI = new CommentAPI(httpManager);
//
// // 测试获取文档评论
// async function testGetDocumentComments() {
//     try {
//         const response = await commentAPI.getDocumentComments({
//             doc_id: '1', // 替换为实际的文档ID
//         });
//         console.log('获取文档评论:', response);
//     } catch (error) {
//         console.error('获取文档评论失败:', error);
//     }
// }
//
// // 测试创建评论
// async function testCreateComment() {
//     try {
//         const response = await commentAPI.createComment({
//             id: 'comment123', // 替换为实际的评论ID
//             doc_id: '1', // 替换为实际的文档ID
//             page_id: 'page1',
//             shape_id: 'shape1',
//             target_shape_id: 'target1',
//             shape_frame: {
//                 x1: 0,
//                 x2: 100,
//                 y1: 0,
//                 y2: 100,
//             },
//             content: '测试评论内容',
//         });
//         console.log('创建评论:', response);
//     } catch (error) {
//         console.error('创建评论失败:', error);
//     }
// }
//
// // 测试编辑评论
// async function testEditComment() {
//     try {
//         const response = await commentAPI.editComment({
//             id: 'comment123', // 替换为实际的评论ID
//             doc_id: '1', // 替换为实际的文档ID
//             page_id: 'page1',
//             shape_id: 'shape1',
//             target_shape_id: 'target1',
//             content: '修改后的评论内容',
//         });
//         console.log('编辑评论:', response);
//     } catch (error) {
//         // console.error('编辑评论失败:', error);
//     }
// }
//
// // 测试删除评论
// async function testDeleteComment() {
//     try {
//         const response = await commentAPI.deleteComment({
//             comment_id: 'comment123', // 替换为实际的评论ID
//             doc_id: '1', // 替换为实际的文档ID
//         });
//         console.log('删除评论:', response);
//     } catch (error) {
//         console.error('删除评论失败:', error);
//     }
// }
//
// // 测试设置评论状态
// async function testSetCommentStatus() {
//     try {
//         const response = await commentAPI.setCommentStatus({
//             id: 'comment123', // 替换为实际的评论ID
//             doc_id: '1', // 替换为实际的文档ID
//             status: 1, // 1表示已解决
//         });
//         console.log('设置评论状态:', response);
//     } catch (error) {
//         console.error('设置评论状态失败:', error);
//     }
// }
//
// // 运行所有测试
// async function runAllTests() {
//     console.log('开始测试评论相关API...');
//
//     await testCreateComment();
//     await testGetDocumentComments();
//     await testEditComment();
//     await testSetCommentStatus();
//     await testDeleteComment();
//
//     console.log('测试完成');
// }
//
// // 执行测试
// runAllTests().catch(console.error);