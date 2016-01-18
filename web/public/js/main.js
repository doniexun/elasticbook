// function doneTyping (t) {
//   $.ajax({
//     method: "POST",
//     url: "/elasticbook/suggest",
//     data: { term: t }
//   })
//   .done(function(res) {
//     console.log("received suggestions");

//     var suggestions = [];
//     var options = res.completion[0]["options"];
//     options.forEach(function(v) {
//       suggestions.push(v['text'])
//     })
//     console.log(suggestions);
//     $("form input[type=text][data-suggest=true]").autocomplete({
//       delay: 300,
//       minLength: 2,
//       autoFocus: true,
//       source: suggestions });
//   })
//   .fail(function() {
//     console.log("error");
//   })
//   .always(function() {
//     console.log("done...")
//   });
// }

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

// $(document).on('keydown', 'form input[type=text]', function(e) {
//   var typingTimer;
//   var doneTypingInterval = 2000;
//   var input = $(this);

//   if ((input.data('suggest') == true)) {
//     input.on('keyup', function () {
//       clearTimeout(typingTimer);
//       var term = input.val()
//       typingTimer = setTimeout(doneTyping(term), doneTypingInterval);
//     });

//     input.on('keydown', function () {
//       clearTimeout(typingTimer);
//     });
//   };
// });
