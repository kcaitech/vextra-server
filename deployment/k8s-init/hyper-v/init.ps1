# 为内部虚拟交换机创建NAT网络，使其能访问外部网络

# 确保管理员权限运行
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
#    Start-Process PowerShell -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File $($MyInvocation.MyCommand)" -Verb RunAs
    Write-Host "please run as administrator"
    exit
}

# 创建NAT实例
New-NetNat -Name KcNat1 -InternalIPInterfaceAddressPrefix 169.254.231.0/24

# 查看NAT实例
# Get-NetNat
# 删除NAT实例
# Remove-NetNat -Name KcNat1
# 查看NAT会话
# Get-NetNatSession
# 查看NAT静态映射
# Get-NetNatStaticMapping
