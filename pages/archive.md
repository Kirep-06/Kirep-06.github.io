---
layout: page
title: 归档
permalink: /archive/
description: 按时间快速浏览历史文章。
---

{% assign years = site.posts | group_by_exp: "post", "post.date | date: '%Y'" %}

<div class="archive-list">
  {% for year in years %}
    <section class="archive-year">
      <h2>{{ year.name }}</h2>
      {% assign months = year.items | group_by_exp: "post", "post.date | date: '%m'" %}
      {% for month in months %}
        <h3>{{ month.name }} 月</h3>
        <ul class="post-list compact">
          {% for post in month.items %}
            <li>
              <time datetime="{{ post.date | date_to_xmlschema }}">{{ post.date | date: "%m-%d" }}</time>
              <a href="{{ post.url | relative_url }}">{{ post.title }}</a>
            </li>
          {% endfor %}
        </ul>
      {% endfor %}
    </section>
  {% endfor %}
</div>
