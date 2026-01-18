// 主要交互逻辑

document.getElementById('registerBtn').addEventListener('click', function() {
  alert('报名功能即将开放，请关注官方通知！');
});

// 可选：添加平滑滚动
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
  anchor.addEventListener('click', function(e) {
    e.preventDefault();
    document.querySelector(this.getAttribute('href')).scrollIntoView({
      behavior: 'smooth'
    });
  });
});