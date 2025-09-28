// 前端API测试脚本
// 用于验证新的API集成是否正常工作

import { getTopicsWithLimit } from '../src/api.js'

// 测试函数
async function testApiIntegration() {
  console.log('🧪 开始测试API集成...')
  
  try {
    console.log('📡 测试获取前10个topics...')
    const response10 = await getTopicsWithLimit(10)
    console.log('✅ 前10个topics:', response10)
    
    console.log('📡 测试获取前30个topics...')
    const response30 = await getTopicsWithLimit(30)
    console.log('✅ 前30个topics:', response30)
    
    console.log('🎉 API集成测试完成！')
  } catch (error) {
    console.error('❌ API测试失败:', error)
  }
}

// 导出测试函数
export { testApiIntegration }

/*
使用方法:
1. 确保后端服务器运行在 http://localhost:8080
2. 确保用户已登录（有有效的JWT token）
3. 在浏览器控制台中运行:
   import('/path/to/test-api.js').then(m => m.testApiIntegration())
*/