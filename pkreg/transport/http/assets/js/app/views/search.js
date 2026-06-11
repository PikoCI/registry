// views/search.js - Home/search page
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  PkReg.Views.SearchView = Backbone.View.extend({
    template: _.template($('#tmpl-search').html()),
    resultTemplate: _.template($('#tmpl-search-result-card').html()),

    events: {
      'click .kind-tab': 'onKindFilter',
      'click .page-link': 'onPageClick'
    },

    initialize: function(options) {
      this.query = options.query || '';
      this.kind = options.kind || '';
      this.page = options.page || 1;
      this.popularTags = [];
      this.totalCount = 0;
      this.types = [];
      this._fetched = false;
      this.fetchResults();
    },

    render: function() {
      this.$el.html(this.template({
        query: this.query,
        kind: this.kind,
        popularTags: this.popularTags
      }));
      if (this._fetched) {
        this.renderResults();
      }
      return this;
    },

    renderResults: function() {
      var $results = this.$('.search-results');
      var $count = this.$('.search-results-count');

      if (this.types.length === 0) {
        $results.html('<div class="empty-state"><i class="bi bi-search"></i><p>No plugins found</p></div>');
        $count.text('');
        this.$('.search-pagination').html('');
        return;
      }

      var self = this;
      $count.text(this.totalCount + ' plugin' + (this.totalCount !== 1 ? 's' : '') + ' found');
      $results.html('');
      _.each(this.types, function(t) {
        $results.append(self.resultTemplate({
          t: t,
          formatKind: PkReg.formatKind,
          kindIcon: PkReg.kindIcon,
          formatDate: PkReg.formatDate
        }));
      });

      this.renderPagination();
    },

    renderPagination: function() {
      var totalPages = Math.ceil(this.totalCount / 20);
      var $pag = this.$('.search-pagination');
      $pag.html('');

      if (totalPages <= 1) return;

      var html = '<nav><ul class="pagination pkreg-pagination justify-content-center">';
      for (var i = 1; i <= totalPages && i <= 10; i++) {
        html += '<li class="page-item' + (i === this.page ? ' active' : '') + '">';
        html += '<a class="page-link" href="#" data-page="' + i + '">' + i + '</a></li>';
      }
      html += '</ul></nav>';
      $pag.html(html);
    },

    fetchResults: function() {
      var self = this;
      var data = { page: this.page, per_page: 20 };
      if (this.query) data.q = this.query;
      if (this.kind) data.kind = this.kind;

      $.ajax({
        url: '/api/plugins',
        method: 'GET',
        data: data,
        dataType: 'json',
        success: function(resp) {
          self._fetched = true;
          self.types = resp.Types || [];
          self.totalCount = resp.TotalCount || 0;
          self.popularTags = resp.TopTags || [];
          self.render();
        }
      });
    },

    onKindFilter: function(e) {
      e.preventDefault();
      var kind = $(e.currentTarget).data('kind') || '';
      var params = [];
      if (this.query) params.push('q=' + encodeURIComponent(this.query));
      if (kind) params.push('kind=' + encodeURIComponent(kind));
      var route = params.length ? 'search?' + params.join('&') : '';
      Backbone.history.navigate(route, { trigger: true });
    },

    onPageClick: function(e) {
      e.preventDefault();
      var page = parseInt($(e.currentTarget).data('page'), 10);
      if (page && page !== this.page) {
        this.page = page;
        this.fetchResults();
      }
    }
  });

})();
