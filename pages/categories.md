---
layout: page
title: 分类
permalink: /categories/
description: 按分类查看文章。
---

{% assign categories = site.categories | sort %}

<div class="taxonomy-list">
  {% for category in categories %}
    <section id="{{ category[0] | uri_escape }}" class="taxonomy-group">
      <h2>{{ category[0] }} <span>({{ category[1].size }})</span></h2>
      <ul class="post-list compact">
        {% for post in category[1] %}
          <li>
            <time datetime="{{ post.date | date_to_xmlschema }}">{{ post.date | date: "%Y-%m-%d" }}</time>
            <a href="{{ post.url | relative_url }}">{{ post.title }}</a>
          </li>
        {% endfor %}
      </ul>
    </section>
  {% endfor %}
</div>
