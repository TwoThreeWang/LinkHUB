
function style_html(html_source, indent_size, indent_character, max_char) {
    //Wrapper function to invoke all the necessary constructors and deal with the output.

      var Parser, multi_parser;

      function Parser() {

        this.pos = 0; //Parser position
        this.token = '';
        this.current_mode = 'CONTENT'; //reflects the current Parser mode: TAG/CONTENT
        this.tags = { //An object to hold tags, their position, and their parent-tags, initiated with default values
          parent: 'parent1',
          parentcount: 1,
          parent1: ''
        };
        this.tag_type = '';
        this.token_text = this.last_token = this.last_text = this.token_type = '';


        this.Utils = { //Uilities made available to the various functions
          whitespace: "\n\r\t ".split(''),
          single_token: 'br,input,link,meta,!doctype,basefont,base,area,hr,wbr,param,img,isindex,?xml,embed'.split(','), //all the single tags for HTML
          extra_liners: 'head,body,/html'.split(','), //for tags that need a line of whitespace before them
          in_array: function (what, arr) {
            for (var i=0; i<arr.length; i++) {
              if (what === arr[i]) {
                return true;
              }
            }
            return false;
          }
        }

        this.get_content = function () { //function to capture regular content between tags

          var char = '';
          var content = [];
          var space = false; //if a space is needed
          while (this.input.charAt(this.pos) !== '<') {
            if (this.pos >= this.input.length) {
              return content.length?content.join(''):['', 'TK_EOF'];
            }

            char = this.input.charAt(this.pos);
            this.pos++;
            this.line_char_count++;


            if (this.Utils.in_array(char, this.Utils.whitespace)) {
              if (content.length) {
                space = true;
              }
              this.line_char_count--;
              continue; //don't want to insert unnecessary space
            }
            else if (space) {
              if (this.line_char_count >= this.max_char) { //insert a line when the max_char is reached
                content.push('\n');
                for (var i=0; i<this.indent_level; i++) {
                  content.push(this.indent_string);
                }
                this.line_char_count = 0;
              }
              else{
                content.push(' ');
                this.line_char_count++;
              }
              space = false;
            }
            content.push(char); //letter at-a-time (or string) inserted to an array
          }
          return content.length?content.join(''):'';
        }

        this.get_script = function () { //get the full content of a script to pass to js_beautify

          var char = '';
          var content = [];
          var reg_match = new RegExp('\<\/script' + '\>', 'igm');
          reg_match.lastIndex = this.pos;
          var reg_array = reg_match.exec(this.input);
          var end_script = reg_array?reg_array.index:this.input.length; //absolute end of script
          while(this.pos < end_script) { //get everything in between the script tags
            if (this.pos >= this.input.length) {
              return content.length?content.join(''):['', 'TK_EOF'];
            }

            char = this.input.charAt(this.pos);
            this.pos++;


            content.push(char);
          }
          return content.length?content.join(''):''; //we might not have any content at all
        }

        this.record_tag = function (tag){ //function to record a tag and its parent in this.tags Object
          if (this.tags[tag + 'count']) { //check for the existence of this tag type
            this.tags[tag + 'count']++;
            this.tags[tag + this.tags[tag + 'count']] = this.indent_level; //and record the present indent level
          }
          else { //otherwise initialize this tag type
            this.tags[tag + 'count'] = 1;
            this.tags[tag + this.tags[tag + 'count']] = this.indent_level; //and record the present indent level
          }
          this.tags[tag + this.tags[tag + 'count'] + 'parent'] = this.tags.parent; //set the parent (i.e. in the case of a div this.tags.div1parent)
          this.tags.parent = tag + this.tags[tag + 'count']; //and make this the current parent (i.e. in the case of a div 'div1')
        }

        this.retrieve_tag = function (tag) { //function to retrieve the opening tag to the corresponding closer
          if (this.tags[tag + 'count']) { //if the openener is not in the Object we ignore it
            var temp_parent = this.tags.parent; //check to see if it's a closable tag.
            while (temp_parent) { //till we reach '' (the initial value);
              if (tag + this.tags[tag + 'count'] === temp_parent) { //if this is it use it
                break;
              }
              temp_parent = this.tags[temp_parent + 'parent']; //otherwise keep on climbing up the DOM Tree
            }
            if (temp_parent) { //if we caught something
              this.indent_level = this.tags[tag + this.tags[tag + 'count']]; //set the indent_level accordingly
              this.tags.parent = this.tags[temp_parent + 'parent']; //and set the current parent
            }
            delete this.tags[tag + this.tags[tag + 'count'] + 'parent']; //delete the closed tags parent reference...
            delete this.tags[tag + this.tags[tag + 'count']]; //...and the tag itself
            if (this.tags[tag + 'count'] == 1) {
              delete this.tags[tag + 'count'];
            }
            else {
              this.tags[tag + 'count']--;
            }
          }
        }

        this.get_tag = function () { //function to get a full tag and parse its type
          var char = '';
          var content = [];
          var space = false;

          do {
            if (this.pos >= this.input.length) {
              return content.length?content.join(''):['', 'TK_EOF'];
            }

            char = this.input.charAt(this.pos);
            this.pos++;
            this.line_char_count++;

            if (this.Utils.in_array(char, this.Utils.whitespace)) { //don't want to insert unnecessary space
              space = true;
              this.line_char_count--;
              continue;
            }

            if (char === "'" || char === '"') {
              if (!content[1] || content[1] !== '!') { //if we're in a comment strings don't get treated specially
                char += this.get_unformatted(char);
                space = true;
              }
            }

            if (char === '=') { //no space before =
              space = false;
            }

            if (content.length && content[content.length-1] !== '=' && char !== '>'
                && space) { //no space after = or before >
              if (this.line_char_count >= this.max_char) {
                this.print_newline(false, content);
                this.line_char_count = 0;
              }
              else {
                content.push(' ');
                this.line_char_count++;
              }
              space = false;
            }
            content.push(char); //inserts character at-a-time (or string)
          } while (char !== '>');

          var tag_complete = content.join('');
          var tag_index;
          if (tag_complete.indexOf(' ') != -1) { //if there's whitespace, thats where the tag name ends
            tag_index = tag_complete.indexOf(' ');
          }
          else { //otherwise go with the tag ending
            tag_index = tag_complete.indexOf('>');
          }
          var tag_check = tag_complete.substring(1, tag_index).toLowerCase();
          if (tag_complete.charAt(tag_complete.length-2) === '/' ||
              this.Utils.in_array(tag_check, this.Utils.single_token)) { //if this tag name is a single tag type (either in the list or has a closing /)
            this.tag_type = 'SINGLE';
          }
          else if (tag_check === 'script') { //for later script handling
            this.record_tag(tag_check);
            this.tag_type = 'SCRIPT';
          }
          else if (tag_check === 'style') { //for future style handling (for now it justs uses get_content)
            this.record_tag(tag_check);
            this.tag_type = 'STYLE';
          }
          else if (tag_check.charAt(0) === '!') { //peek for <!-- comment
            if (tag_check.indexOf('[if') != -1) { //peek for <!--[if conditional comment
              if (tag_complete.indexOf('!IE') != -1) { //this type needs a closing --> so...
                var comment = this.get_unformatted('-->', tag_complete); //...delegate to get_unformatted
                content.push(comment);
              }
              this.tag_type = 'START';
            }
            else if (tag_check.indexOf('[endif') != -1) {//peek for <!--[endif end conditional comment
              this.tag_type = 'END';
              this.unindent();
            }
            else if (tag_check.indexOf('[cdata[') != -1) { //if it's a <[cdata[ comment...
              var comment = this.get_unformatted(']]>', tag_complete); //...delegate to get_unformatted function
              content.push(comment);
              this.tag_type = 'SINGLE'; //<![CDATA[ comments are treated like single tags
            }
            else {
              var comment = this.get_unformatted('-->', tag_complete);
              content.push(comment);
              this.tag_type = 'SINGLE';
            }
          }
          else {
            if (tag_check.charAt(0) === '/') { //this tag is a double tag so check for tag-ending
              this.retrieve_tag(tag_check.substring(1)); //remove it and all ancestors
              this.tag_type = 'END';
            }
            else { //otherwise it's a start-tag
              this.record_tag(tag_check); //push it on the tag stack
              this.tag_type = 'START';
            }
            if (this.Utils.in_array(tag_check, this.Utils.extra_liners)) { //check if this double needs an extra line
              this.print_newline(true, this.output);
            }
          }
          return content.join(''); //returns fully formatted tag
        }

        this.get_unformatted = function (delimiter, orig_tag) { //function to return unformatted content in its entirety

          if (orig_tag && orig_tag.indexOf(delimiter) != -1) {
            return '';
          }
          var char = '';
          var content = '';
          var space = true;
          do {


            char = this.input.charAt(this.pos);
            this.pos++

            if (this.Utils.in_array(char, this.Utils.whitespace)) {
              if (!space) {
                this.line_char_count--;
                continue;
              }
              if (char === '\n' || char === '\r') {
                content += '\n';
                for (var i=0; i<this.indent_level; i++) {
                  content += this.indent_string;
                }
                space = false; //...and make sure other indentation is erased
                this.line_char_count = 0;
                continue;
              }
            }
            content += char;
            this.line_char_count++;
            space = true;


          } while (content.indexOf(delimiter) == -1);
          return content;
        }

        this.get_token = function () { //initial handler for token-retrieval
          var token;

          if (this.last_token === 'TK_TAG_SCRIPT') { //check if we need to format javascript
            var temp_token = this.get_script();
            if (typeof temp_token !== 'string') {
              return temp_token;
            }
            token = js_beautify(temp_token, this.indent_size, this.indent_character, this.indent_level); //call the JS Beautifier
            return [token, 'TK_CONTENT'];
          }
          if (this.current_mode === 'CONTENT') {
            token = this.get_content();
            if (typeof token !== 'string') {
              return token;
            }
            else {
              return [token, 'TK_CONTENT'];
            }
          }

          if(this.current_mode === 'TAG') {
            token = this.get_tag();
            if (typeof token !== 'string') {
              return token;
            }
            else {
              var tag_name_type = 'TK_TAG_' + this.tag_type;
              return [token, tag_name_type];
            }
          }
        }

        this.printer = function (js_source, indent_character, indent_size, max_char) { //handles input/output and some other printing functions

          this.input = js_source || ''; //gets the input for the Parser
          this.output = [];
          this.indent_character = indent_character || ' ';
          this.indent_string = '';
          this.indent_size = indent_size || 2;
          this.indent_level = 0;
          this.max_char = max_char || 70; //maximum amount of characters per line
          this.line_char_count = 0; //count to see if max_char was exceeded

          for (var i=0; i<this.indent_size; i++) {
            this.indent_string += this.indent_character;
          }

          this.print_newline = function (ignore, arr) {
            this.line_char_count = 0;
            if (!arr || !arr.length) {
              return;
            }
            if (!ignore) { //we might want the extra line
              while (this.Utils.in_array(arr[arr.length-1], this.Utils.whitespace)) {
                arr.pop();
              }
            }
            arr.push('\n');
            for (var i=0; i<this.indent_level; i++) {
              arr.push(this.indent_string);
            }
          }


          this.print_token = function (text) {
            this.output.push(text);
          }

          this.indent = function () {
            this.indent_level++;
          }

          this.unindent = function () {
            if (this.indent_level > 0) {
              this.indent_level--;
            }
          }
        }
        return this;
      }

      /*_____________________--------------------_____________________*/



      multi_parser = new Parser(); //wrapping functions Parser
      multi_parser.printer(html_source, indent_character, indent_size); //initialize starting values



      while (true) {
          var t = multi_parser.get_token();
          multi_parser.token_text = t[0];
          multi_parser.token_type = t[1];

        if (multi_parser.token_type === 'TK_EOF') {
          break;
        }


        switch (multi_parser.token_type) {
          case 'TK_TAG_START': case 'TK_TAG_SCRIPT': case 'TK_TAG_STYLE':
            multi_parser.print_newline(false, multi_parser.output);
            multi_parser.print_token(multi_parser.token_text);
            multi_parser.indent();
            multi_parser.current_mode = 'CONTENT';
            break;
          case 'TK_TAG_END':
            multi_parser.print_newline(true, multi_parser.output);
            multi_parser.print_token(multi_parser.token_text);
            multi_parser.current_mode = 'CONTENT';
            break;
          case 'TK_TAG_SINGLE':
            multi_parser.print_newline(false, multi_parser.output);
            multi_parser.print_token(multi_parser.token_text);
            multi_parser.current_mode = 'CONTENT';
            break;
          case 'TK_CONTENT':
            if (multi_parser.token_text !== '') {
              multi_parser.print_newline(false, multi_parser.output);
              multi_parser.print_token(multi_parser.token_text);
            }
            multi_parser.current_mode = 'TAG';
            break;
        }
        multi_parser.last_token = multi_parser.token_type;
        multi_parser.last_text = multi_parser.token_text;
      }
      return multi_parser.output.join('');
    }