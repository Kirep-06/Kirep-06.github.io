---
layout: page
title: 标签
permalink: /tags/
description: 按标签查看文章。
---

{% assign tags = site.tags | sort %}

<div class="taxonomy-list">
  {% for tag in tags %}
    <section id="{{ tag[0] | uri_escape }}" class="taxonomy-group">
      <h2>#{{ tag[0] }} <span>({{ tag[1].size }})</span></h2>
      <ul class="post-list compact">
        {% for post in tag[1] %}
          <li>
            <time datetime="{{ post.date | date_to_xmlschema }}">{{ post.date | date: "%Y-%m-%d" }}</time>
            <a href="{{ post.url | relative_url }}">{{ post.title }}</a>
          </li>
        {% endfor %}
      </ul>
    </section>
  {% endfor %}
</div>
