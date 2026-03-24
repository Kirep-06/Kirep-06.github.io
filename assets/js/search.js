(function () {
  var input = document.getElementById('search-input');
  var list = document.getElementById('search-results');
  var status = document.getElementById('search-status');
  var posts = [];

  if (!input || !list || !status) return;

  function escapeHtml(text) {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function render(items, query) {
    list.innerHTML = '';

    if (!query) {
      status.textContent = '输入关键词开始搜索。';
      return;
    }

    if (!items.length) {
      status.textContent = '没有找到匹配内容。';
      return;
    }

    status.textContent = '找到 ' + items.length + ' 条结果。';

    items.forEach(function (post) {
      var li = document.createElement('li');
      li.innerHTML =
        '<h3><a href="' + post.url + '">' + escapeHtml(post.title) + '</a></h3>' +
        '<p class="post-meta">' + escapeHtml(post.date) + '</p>' +
        '<p class="post-excerpt">' + escapeHtml((post.excerpt || '').slice(0, 120)) + '</p>';
      list.appendChild(li);
    });
  }

  function search() {
    var query = input.value.trim().toLowerCase();
    if (!query) {
      render([], '');
      return;
    }

    var terms = query.split(/\s+/).filter(Boolean);
    var results = posts.filter(function (post) {
      var haystack = [
        post.title || '',
        post.excerpt || '',
        post.content || '',
        (post.tags || []).join(' '),
        (post.categories || []).join(' ')
      ]
        .join(' ')
        .toLowerCase();

      return terms.every(function (term) {
        return haystack.indexOf(term) !== -1;
      });
    });

    render(results, query);
  }

  var baseUrl = document.body.getAttribute('data-baseurl') || '';
  var searchUrl = (baseUrl + '/search.json').replace(/\/{2,}/g, '/');

  fetch(searchUrl)
    .then(function (res) {
      return res.json();
    })
    .then(function (data) {
      posts = data || [];
      status.textContent = '索引已加载，可搜索标题、摘要与正文。';
    })
    .catch(function () {
      status.textContent = '搜索索引加载失败，请稍后刷新重试。';
    });

  input.addEventListener('input', search);
})();
