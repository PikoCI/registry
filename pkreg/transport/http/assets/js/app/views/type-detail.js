// views/type-detail.js - Type detail page
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  PkReg.Views.TypeDetailView = Backbone.View.extend({
    template: _.template($('#tmpl-type-detail').html()),

    events: {
      'click .detail-tab': 'onTabClick',
      'change .version-select': 'onVersionChange',
      'click .btn-yank': 'onYank',
      'click .btn-deprecate': 'onDeprecate',
      'click .btn-delete-version': 'onDeleteVersion',
      'click .btn-delete-type': 'onDeleteType',
      'click .btn-remove-tag': 'onRemoveTag',
      'keydown .tag-add-input': 'onAddTag'
    },

    initialize: function(options) {
      this.namespace = options.namespace;
      this.name = options.name;
      this.activeTab = 'readme';
      this.selectedVersionIdx = 0;
      this.data = null;
      this.fetchData();
    },

    render: function() {
      if (!this.data) {
        this.$el.html('<div class="loading-spinner"><div class="spinner-border text-primary" role="status"><span class="visually-hidden">Loading...</span></div></div>');
        return this;
      }

      var versions = this.data.versions || [];
      var selectedVersion = versions[this.selectedVersionIdx] || {};
      var params = this._extractParams(selectedVersion);

      var user = PkReg.Auth.getUser();
      var isOwner = user && user.Username === this.namespace;

      this.$el.html(this.template({
        type: this.data.type,
        versions: versions,
        selectedVersion: selectedVersion,
        tags: this.data.tags || [],
        params: params,
        activeTab: this.activeTab,
        namespace: this.namespace,
        isOwner: isOwner,
        formatKind: PkReg.formatKind,
        kindIcon: PkReg.kindIcon,
        formatDate: PkReg.formatDate
      }));

      // Render markdown readme
      if (this.activeTab === 'readme' && selectedVersion.Readme) {
        var html = DOMPurify.sanitize(marked.parse(selectedVersion.Readme));
        this.$('.readme-rendered').html(html);
      }

      return this;
    },

    _extractParams: function(version) {
      var params = [];
      if (!version) return params;

      if (version.Params) {
        try {
          params = typeof version.Params === 'string' ? JSON.parse(version.Params) : version.Params;
        } catch(e) {}
        if (!Array.isArray(params)) params = [];
      }

      if (params.length === 0 && version.Content) {
        var match = version.Content.match(/params\s*=\s*\[([\s\S]*?)\]/);
        if (match) {
          var raw = match[1];
          var re = /"([^"]+)"/g;
          var m;
          while ((m = re.exec(raw)) !== null) {
            params.push({ Name: m[1] });
          }
        }
      }

      return params;
    },

    fetchData: function() {
      var self = this;
      $.ajax({
        url: '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name),
        method: 'GET',
        dataType: 'json',
        success: function(resp) {
          self.data = resp;
          self.render();
        },
        error: function() {
          self.$el.html('<div class="empty-state"><i class="bi bi-exclamation-triangle"></i><p>Plugin not found</p><a href="#" class="btn btn-outline-primary">Back to Home</a></div>');
        }
      });
    },

    _apiCall: function(method, path, body, cb) {
      var opts = {
        url: path,
        method: method,
        headers: PkReg.Auth.ajaxHeaders()
      };
      if (body) {
        opts.contentType = 'application/json';
        opts.data = JSON.stringify(body);
      }
      opts.success = cb;
      opts.error = function(xhr) {
        var msg = 'Action failed';
        try { msg = JSON.parse(xhr.responseText).error; } catch(e) {}
        alert(msg);
      };
      $.ajax(opts);
    },

    onTabClick: function(e) {
      e.preventDefault();
      var tab = $(e.currentTarget).data('tab');
      if (tab && tab !== this.activeTab) {
        this.activeTab = tab;
        this.render();
      }
    },

    onVersionChange: function(e) {
      var selected = $(e.currentTarget).val();
      var versions = this.data.versions || [];
      for (var i = 0; i < versions.length; i++) {
        if (versions[i].Version === selected) {
          this.selectedVersionIdx = i;
          break;
        }
      }
      this.render();
    },

    onYank: function(e) {
      e.preventDefault();
      var ver = this.data.versions[this.selectedVersionIdx];
      if (!ver) return;
      if (!confirm('Yank version ' + ver.Version + '? It will still be visible but marked as yanked.')) return;

      var self = this;
      var path = '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name) + '/' + encodeURIComponent(ver.Version) + '/yank';
      this._apiCall('POST', path, null, function() {
        self.fetchData();
      });
    },

    onDeprecate: function(e) {
      e.preventDefault();
      var ver = this.data.versions[this.selectedVersionIdx];
      if (!ver) return;

      var message = prompt('Deprecation message (optional):');
      if (message === null) return; // cancelled
      var successor = prompt('Successor version (optional):');
      if (successor === null) successor = '';

      var self = this;
      var path = '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name) + '/' + encodeURIComponent(ver.Version) + '/deprecate';
      this._apiCall('POST', path, { message: message, successor: successor }, function() {
        self.fetchData();
      });
    },

    onDeleteVersion: function(e) {
      e.preventDefault();
      var ver = this.data.versions[this.selectedVersionIdx];
      if (!ver) return;
      if (!confirm('Delete version ' + ver.Version + '? This cannot be undone.')) return;

      var self = this;
      var path = '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name) + '/' + encodeURIComponent(ver.Version);
      this._apiCall('DELETE', path, null, function() {
        self.selectedVersionIdx = 0;
        self.fetchData();
      });
    },

    onRemoveTag: function(e) {
      e.preventDefault();
      e.stopPropagation();
      var tagToRemove = $(e.currentTarget).data('tag');
      var tags = _.without(this.data.tags || [], tagToRemove);
      this._saveTags(tags);
    },

    onAddTag: function(e) {
      if (e.key !== 'Enter') return;
      e.preventDefault();
      var input = this.$('.tag-add-input');
      var newTag = input.val().trim().toLowerCase();
      if (!newTag) return;

      var tags = (this.data.tags || []).slice();
      if (tags.indexOf(newTag) === -1) {
        tags.push(newTag);
      }
      this._saveTags(tags);
    },

    _saveTags: function(tags) {
      var self = this;
      var path = '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name) + '/tags';
      this._apiCall('PATCH', path, { tags: tags }, function() {
        self.fetchData();
      });
    },

    onDeleteType: function(e) {
      e.preventDefault();
      if (!confirm('Delete plugin "' + this.name + '" and ALL its versions? This cannot be undone.')) return;

      var self = this;
      var path = '/api/plugins/' + encodeURIComponent(this.namespace) + '/' + encodeURIComponent(this.name);
      this._apiCall('DELETE', path, null, function() {
        Backbone.history.navigate('me/types', { trigger: true });
      });
    }
  });

})();
