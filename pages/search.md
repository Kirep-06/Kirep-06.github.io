---
layout: page
title: 搜索
permalink: /search/
description: 本地全文搜索（标题、摘要、正文关键词）。
scripts:
  - /assets/js/search.js
---

<div class="search-box">
  <label for="search-input">关键词</label>
  <input id="search-input" type="search" placeholder="输入关键词，例如：Jekyll、前端、工具链" autocomplete="off">
</div>

<p id="search-status" class="search-status">输入关键词开始搜索。</p>
<ul id="search-results" class="post-list"></ul>
