import { ShareAPI } from '../request/share'
import { HttpMgr } from '../request/http'
import { TEST_API_URL, TEST_TOKEN, TEST_UNAUTHORIZED } from './config';

// 创建HTTP管理器实例
const httpManager = new HttpMgr(TEST_API_URL, TEST_UNAUTHORIZED, TEST_TOKEN);

// 创建ShareAPI实例
const shareAPI = new ShareAPI(httpManager);

// 测试获取分享列表
async function testGetShareList() {
    try {
        const response = await shareAPI.getShareListAPI({
            page: 1,
            page_size: 10,
        });
        console.log('获取分享列表:', response);
    } catch (error) {
        console.error('获取分享列表失败:', error);
    }
}

// 测试获取文档权限
async function testGetDocumentAuthority() {
    try {
        const response = await shareAPI.getDocumentAuthorityAPI({
            document_id: '123', // 替换为实际的文档ID
        });
        console.log('获取文档权限:', response);
    } catch (error) {
        console.error('获取文档权限失败:', error);
    }
}

// 测试获取文档密钥
async function testGetDocumentKey() {
    try {
        const response = await shareAPI.getDocumentKeyAPI({
            document_id: '123', // 替换为实际的文档ID
        });
        console.log('获取文档密钥:', response);
    } catch (error) {
        console.error('获取文档密钥失败:', error);
    }
}

// 测试获取文档信息
async function testGetDocumentInfo() {
    try {
        const response = await shareAPI.getDocumentInfoAPI({
            document_id: '123', // 替换为实际的文档ID
        });
        console.log('获取文档信息:', response);
    } catch (error) {
        console.error('获取文档信息失败:', error);
    }
}

// 测试设置分享类型
async function testSetShareType() {
    try {
        const response = await shareAPI.setShateTypeAPI({
            document_id: '123', // 替换为实际的文档ID
            share_type: 'public', // public表示公开分享
        });
        console.log('设置分享类型:', response);
    } catch (error) {
        console.error('设置分享类型失败:', error);
    }
}

// 测试更新分享权限
async function testPutShareAuthority() {
    try {
        const response = await shareAPI.putShareAuthorityAPI({
            document_id: '123', // 替换为实际的文档ID
            permissions: ['read'], // 只读权限
        });
        console.log('更新分享权限:', response);
    } catch (error) {
        console.error('更新分享权限失败:', error);
    }
}

// 测试删除分享权限
async function testDelShareAuthority() {
    try {
        const response = await shareAPI.delShareAuthorityAPI({
            document_id: '123', // 替换为实际的文档ID
        });
        console.log('删除分享权限:', response);
    } catch (error) {
        console.error('删除分享权限失败:', error);
    }
}

// 测试申请文档权限
async function testPostDocumentAuthority() {
    try {
        const response = await shareAPI.postDocumentAuthorityAPI({
            document_id: '123', // 替换为实际的文档ID
            permissions: ['read'], // 申请只读权限
            reason: '需要查看文档内容',
        });
        console.log('申请文档权限:', response);
    } catch (error) {
        console.error('申请文档权限失败:', error);
    }
}

// 测试获取申请列表
async function testGetApplyList() {
    try {
        const response = await shareAPI.getApplyListAPI({
            page: 1,
            page_size: 10,
        });
        console.log('获取申请列表:', response);
    } catch (error) {
        console.error('获取申请列表失败:', error);
    }
}

// 测试权限申请审核
async function testPromissionApplyAudit() {
    try {
        const response = await shareAPI.promissionApplyAuditAPI({
            apply_id: 'apply123', // 替换为实际的申请ID
            status: 'approved', // 通过申请
            reason: '同意申请',
        });
        console.log('权限申请审核:', response);
    } catch (error) {
        console.error('权限申请审核失败:', error);
    }
}

// 运行所有测试
async function runAllTests() {
    console.log('开始测试分享相关API...');
    
    await testGetShareList();
    await testGetDocumentAuthority();
    await testGetDocumentKey();
    await testGetDocumentInfo();
    await testSetShareType();
    await testPutShareAuthority();
    await testDelShareAuthority();
    await testPostDocumentAuthority();
    await testGetApplyList();
    await testPromissionApplyAudit();
    
    console.log('测试完成');
}

// 执行测试
runAllTests().catch(console.error); 