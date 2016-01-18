$(document).ready(function() {
  if (window.location.href.indexOf("/search") == -1) {
    $("form input[type=text][data-suggest=true]").get(0).focus();
  }
});

$(document).ready(function() {
  $("form input[type=text][data-suggest=true]").autocomplete({
      delay: 300,
      minLength: 2,
      autoFocus: true,
      source: function( request, response ) {
        $.ajax({
          method: "POST",
          dataType: "json",
          url: "/elasticbook/suggest",
          data: { term: request.term },
          success: function( data ) {
            var suggestions = [];
            var options = data.completion[0]["options"];
            options.forEach(function(v) {
              suggestions.push(v['text'])
            })
            console.log(suggestions);
            response( suggestions );
          }
        });
      },
      select: function( event, ui ) {
        console.log( ui.item ?
          "Selected: " + ui.item.label :
          "Nothing selected, input was " + this.value);
      },
      open: function() {
        // $( this ).removeClass( "ui-corner-all" ).addClass( "ui-corner-top" );
      },
      close: function() {
        // $( this ).removeClass( "ui-corner-top" ).addClass( "ui-corner-all" );
      }
    });
});
