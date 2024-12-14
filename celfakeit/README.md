## celfakeit
This is a small library which exposes many [gofakeit](https://github.com/brianvoe/gofakeit) functions to CEL. All functions are simple wrappers to [functions in gofakeit](https://github.com/brianvoe/gofakeit?tab=readme-ov-file#functions).

```go
import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/sudorandom/fauxrpc/celfakeit"
)

...

env, err := cel.NewEnv(celfakeit.Configure())
if err != nil {
    // handle error
}

ast, issues := env.Compile(`fake_book_author()`)
if err != nil {
    // handle error
}

program, err := env.Program(ast)
if err != nil {
    // handle error
}

out, _, err := program.Eval(map[string]any{})
if err != nil {
    // handle error
}
fmt.Println(out)

```

### Functions:

- `fake_file_extension`
- `fake_product_name`
- `fake_product_description`
- `fake_product_category`
- `fake_product_feature`
- `fake_product_material`
- `fake_product_upc`
- `fake_product_audience`
- `fake_product_dimension`
- `fake_product_usecase`
- `fake_product_benefit`
- `fake_product_suffix`
- `fake_name`
- `fake_name_prefix`
- `fake_name_suffix`
- `fake_first_name`
- `fake_middle_name`
- `fake_last_name`
- `fake_gender`
- `fake_ssn`
- `fake_hobby`
- `fake_email`
- `fake_phone`
- `fake_phone_formatted`
- `fake_username`
- `fake_city`
- `fake_country`
- `fake_country_abr`
- `fake_state`
- `fake_state_abr`
- `fake_street`
- `fake_street_name`
- `fake_street_number`
- `fake_street_prefix`
- `fake_street_suffix`
- `fake_zip`
- `fake_latitude`
- `fake_longitude`
- `fake_gamertag`
- `fake_beer_alcohol`
- `fake_beer_blg`
- `fake_beer_hop`
- `fake_beer_ibu`
- `fake_beer_malt`
- `fake_beer_name`
- `fake_beer_style`
- `fake_beer_yeast`
- `fake_car_maker`
- `fake_car_model`
- `fake_car_type`
- `fake_car_transmission_type`
- `fake_noun`
- `fake_noun_common`
- `fake_noun_concrete`
- `fake_noun`
- `fake_noun_common`
- `fake_noun_concrete`
- `fake_noun_abstract`
- `fake_noun_collective_people`
- `fake_noun_collective_animal`
- `fake_noun_collective_thing`
- `fake_noun_countable`
- `fake_noun_uncountable`
- `fake_verb`
- `fake_verb_action`
- `fake_verb_linking`
- `fake_verb_helping`
- `fake_adverb`
- `fake_adverb_manner`
- `fake_adverb_degree`
- `fake_adverb_place`
- `fake_adverb_time_definite`
- `fake_adverb_time_indefinite`
- `fake_adverb_frequency_definite`
- `fake_adverb_frequency_indefinite`
- `fake_preposition`
- `fake_preposition_simple`
- `fake_preposition_double`
- `fake_preposition_compound`
- `fake_adjective`
- `fake_adjective_descriptive`
- `fake_adjective_quantitative`
- `fake_adjective_proper`
- `fake_adjective_demonstrative`
- `fake_adjective_possessive`
- `fake_adjective_interrogative`
- `fake_adjective_indefinite`
- `fake_pronoun`
- `fake_pronoun_personal`
- `fake_pronoun_object`
- `fake_pronoun_possessive`
- `fake_pronoun_reflective`
- `fake_pronoun_demonstrative`
- `fake_pronoun_interrogative`
- `fake_pronoun_relative`
- `fake_connective`
- `fake_connective_time`
- `fake_connective_comparative`
- `fake_connective_complaint`
- `fake_connective_listing`
- `fake_connective_casual`
- `fake_connective_examplify`
- `fake_lorem_ipsum_word`
- `fake_question`
- `fake_quote`
- `fake_phrase`
- `fake_fruit`
- `fake_vegetable`
- `fake_breakfast`
- `fake_lunch`
- `fake_dinner`
- `fake_snack`
- `fake_dessert`
- `fake_bool`
- `fake_uuid`
- `fake_flip_a_coin`
- `fake_color`
- `fake_hex_color`
- `fake_safe_color`
- `fake_url`
- `fake_domain_name`
- `fake_domain_suffix`
- `fake_i_pv4_address`
- `fake_i_pv6_address`
- `fake_mac_address`
- `fake_http_status_code`
- `fake_http_status_code_simple`
- `fake_http_method`
- `fake_http_version`
- `fake_user_agent`
- `fake_chrome_user_agent`
- `fake_firefox_user_agent`
- `fake_opera_user_agent`
- `fake_safari_user_agent`
- `fake_input_name`
- `fake_date`
- `fake_past_date`
- `fake_future_date`
- `fake_nanosecond`
- `fake_second`
- `fake_minute`
- `fake_hour`
- `fake_month`
- `fake_month_string`
- `fake_day`
- `fake_week_day`
- `fake_year`
- `fake_time_zone`
- `fake_time_zone_abv`
- `fake_time_zone_full`
- `fake_time_offset`
- `fake_time_zone_region`
- `fake_credit_card_cvv`
- `fake_credit_card_exp`
- `fake_credit_card_type`
- `fake_currency_long`
- `fake_currency_short`
- `fake_ach_routing`
- `fake_ach_account`
- `fake_bitcoin_address`
- `fake_bitcoin_private_key`
- `fake_cusip`
- `fake_isin`
- `fake_bs`
- `fake_blurb`
- `fake_buzz_word`
- `fake_company`
- `fake_company_suffix`
- `fake_job_descriptor`
- `fake_job_level`
- `fake_job_title`
- `fake_slogan`
- `fake_hacker_abbreviation`
- `fake_hacker_adjective`
- `fake_hackering_verb`
- `fake_hacker_noun`
- `fake_hacker_phrase`
- `fake_hacker_verb`
- `fake_hipster_word`
- `fake_app_name`
- `fake_app_version`
- `fake_app_author`
- `fake_pet_name`
- `fake_animal`
- `fake_animal_type`
- `fake_farm_animal`
- `fake_cat`
- `fake_dog`
- `fake_bird`
- `fake_emoji`
- `fake_emoji_description`
- `fake_emoji_category`
- `fake_emoji_alias`
- `fake_emoji_tag`
- `fake_language`
- `fake_language_abbreviation`
- `fake_programming_language`
- `fake_programming_language_best`
- `fake_int`
- `fake_intn`
- `fake_int8`
- `fake_int16`
- `fake_int32`
- `fake_int64`
- `fake_uint`
- `fake_uintn`
- `fake_uint8`
- `fake_uint16`
- `fake_uint32`
- `fake_uint64`
- `fake_float32`
- `fake_float64`
- `fake_digit`
- `fake_letter`
- `fake_celebrity_actor`
- `fake_celebrity_business`
- `fake_celebrity_sport`
- `fake_minecraft_ore`
- `fake_minecraft_wood`
- `fake_minecraft_armor_tier`
- `fake_minecraft_armor_part`
- `fake_minecraft_weapon`
- `fake_minecraft_tool`
- `fake_minecraft_dye`
- `fake_minecraft_food`
- `fake_minecraft_animal`
- `fake_minecraft_villager_job`
- `fake_minecraft_villager_station`
- `fake_minecraft_villager_level`
- `fake_minecraft_mob_passive`
- `fake_minecraft_mob_neutral`
- `fake_minecraft_mob_hostile`
- `fake_minecraft_mob_boss`
- `fake_minecraft_biome`
- `fake_minecraft_weather`
- `fake_book_title`
- `fake_book_author`
- `fake_book_genre`
- `fake_movie_name`
- `fake_movie_genre`
- `fake_school`
