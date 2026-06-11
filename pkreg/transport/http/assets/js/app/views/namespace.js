// views/namespace.js - Namespace page
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  PkReg.Views.NamespaceView = Backbone.View.extend({
    template: _.template($('#tmpl-namespace').html()),
    cardTemplate: _.template($('#tmpl-search-result-card').html()),

    initialize: function(options) {
      this.namespace = options.namespace;
      this.types = [];
      this.fetchTypes();
    },

    render: function() {
      this.$el.html(this.template({
        namespace: this.namespace,
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
        url: '/api/plugins/' + encodeURIComponent(this.namespace),
        method: 'GET',
        dataType: 'json',
        success: function(resp) {
          self.types = resp || [];
          self.render();
        },
        error: function() {
          self.$el.html('<div class="empty-state"><i class="bi bi-person-x"></i><p>Namespace not found</p><a href="#" class="btn btn-outline-primary">Back to Home</a></div>');
        }
      });
    }
  });

})();
