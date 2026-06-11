// views/tags.js - Tags page
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  // Tag cloud page
  PkReg.Views.TagsView = Backbone.View.extend({
    template: _.template($('#tmpl-tags').html()),

    initialize: function() {
      this.tags = [];
      this.fetchTags();
    },

    render: function() {
      this.$el.html(this.template({ tags: this.tags }));
      return this;
    },

    fetchTags: function() {
      var self = this;
      $.ajax({
        url: '/api/tags',
        method: 'GET',
        dataType: 'json',
        success: function(resp) {
          self.tags = resp || [];
          self.render();
        },
        error: function() {
          self.tags = [];
          self.render();
        }
      });
    }
  });

  // Tag detail - types for a specific tag
  PkReg.Views.TagDetailView = Backbone.View.extend({
    template: _.template($('#tmpl-tag-detail').html()),
    cardTemplate: _.template($('#tmpl-search-result-card').html()),

    initialize: function(options) {
      this.tagName = options.tagName;
      this.types = [];
      this.fetchTypes();
    },

    render: function() {
      this.$el.html(this.template({
        tagName: this.tagName,
        types: this.types,
        formatKind: PkReg.formatKind,
        kindIcon: PkReg.kindIcon,
        formatDate: PkReg.formatDate
      }));

      var self = this;
      var $grid = this.$('.type-cards-grid');
      _.each(this.types, function(t) {
        $grid.append(self.cardTemplate({
          t: t,
          formatKind: PkReg.formatKind,
          kindIcon: PkReg.kindIcon,
          formatDate: PkReg.formatDate
        }));
      });

      return this;
    },

    fetchTypes: function() {
      var self = this;
      $.ajax({
        url: '/api/tags/' + encodeURIComponent(this.tagName),
        method: 'GET',
        dataType: 'json',
        success: function(resp) {
          self.types = resp || [];
          self.render();
        },
        error: function() {
          self.$el.html('<div class="empty-state"><i class="bi bi-tag"></i><p>Tag not found</p></div>');
        }
      });
    }
  });

})();
