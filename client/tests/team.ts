import { TeamAPI } from '../request/team';
import { HttpMgr } from '../request/http';
import { TeamPermType } from '../request/team';
import { TEST_API_URL, TEST_TOKEN, TEST_UNAUTHORIZED } from './config';

// 创建HTTP管理器实例
const httpManager = new HttpMgr(TEST_API_URL, TEST_UNAUTHORIZED, TEST_TOKEN);

// 创建TeamAPI实例
const teamAPI = new TeamAPI(httpManager);

// 测试创建团队
async function testCreateTeam() {
  try {
    const response = await teamAPI.createTeam({
      name: '测试团队',
      description: '这是一个测试团队',
    });
    console.log('创建团队:', response);
  } catch (error) {
    console.error('创建团队失败:', error);
  }
}

// 测试获取团队列表
async function testGetTeamList() {
  try {
    const response = await teamAPI.getTeamList({
      page: 1,
      page_size: 10,
    });
    console.log('获取团队列表:', response);
  } catch (error) {
    console.error('获取团队列表失败:', error);
  }
}

// 测试获取团队详情
async function testGetTeamInfo() {
  try {
    const response = await teamAPI.getTeamInfo({
      team_id: '123', // 替换为实际的团队ID
    });
    console.log('获取团队详情:', response);
  } catch (error) {
    console.error('获取团队详情失败:', error);
  }
}

// 测试更新团队信息
async function testSetTeamInfo() {
  try {
    const response = await teamAPI.setTeamInfo({
      team_id: '123', // 替换为实际的团队ID
      name: '更新后的团队名称',
      description: '更新后的团队描述',
    });
    console.log('更新团队信息:', response);
  } catch (error) {
    console.error('更新团队信息失败:', error);
  }
}

// 测试删除团队
async function testDeleteTeam() {
  try {
    const response = await teamAPI.deleteTeam({
      team_id: '123', // 替换为实际的团队ID
    });
    console.log('删除团队:', response);
  } catch (error) {
    console.error('删除团队失败:', error);
  }
}

// 测试设置团队成员权限
async function testSetTeamMemberPermission() {
  try {
    const response = await teamAPI.setTeamMemberPermission({
      team_id: '123', // 替换为实际的团队ID
      user_id: 'user123', // 替换为实际的用户ID
      perm_type: TeamPermType.Admin, // 设置为管理员权限
    });
    console.log('设置团队成员权限:', response);
  } catch (error) {
    console.error('设置团队成员权限失败:', error);
  }
}

// 测试设置团队成员昵称
async function testSetTeamMemberNickname() {
  try {
    const response = await teamAPI.setTeamMemberNickname({
      team_id: '123', // 替换为实际的团队ID
      user_id: 'user123', // 替换为实际的用户ID
      nickname: '新昵称',
    });
    console.log('设置团队成员昵称:', response);
  } catch (error) {
    console.error('设置团队成员昵称失败:', error);
  }
}

// 测试退出团队
async function testExitTeam() {
  try {
    const response = await teamAPI.exitTeam({
      team_id: '123', // 替换为实际的团队ID
    });
    console.log('退出团队:', response);
  } catch (error) {
    console.error('退出团队失败:', error);
  }
}

// 测试获取团队成员列表
async function testGetTeamMemberList() {
  try {
    const response = await teamAPI.getTeamMemberList({
      team_id: '123', // 替换为实际的团队ID
      page: 1,
      page_size: 10,
    });
    console.log('获取团队成员列表:', response);
  } catch (error) {
    console.error('获取团队成员列表失败:', error);
  }
}

// 运行所有测试
async function runAllTests() {
  console.log('开始测试团队相关API...');
  
  await testCreateTeam();
  await testGetTeamList();
  await testGetTeamInfo();
  await testSetTeamInfo();
  await testDeleteTeam();
  await testSetTeamMemberPermission();
  await testSetTeamMemberNickname();
  await testExitTeam();
  await testGetTeamMemberList();
  
  console.log('测试完成');
}

// 执行测试
runAllTests().catch(console.error); 