import Modal from '@/components/Modal'
import { shallowMount } from '@vue/test-utils'
import { expect } from 'chai'

describe('Modal', () => {

  it('has correct props and attributes', () => {

    const movieTitle = 'Star Wars';
    const movieId = 1;
    const cityId = 'moskow';
    const cityName = 'Moscow';

    const cinemas = [
      {
        'id': 1,
        'name': 'Сатурн',
        'seances': [
          {
            'id': 1,
            'price': 350,
            'seats': [ 
              {'id': 1, 'vacant': true}, 
              {'id': 2, 'vacant': true},
              {'id': 3, 'vacant': false}, 
              {'id': 4, 'vacant': true},
              {'id': 5, 'vacant': true}, 
              {'id': 6, 'vacant': true},
              {'id': 7, 'vacant': false}, 
              {'id': 8, 'vacant': true},
              {'id': 9, 'vacant': false}, 
              {'id': 10, 'vacant': true}
            ],
            'date': '2020-04-23T15:25:43.511Z'
          }     
        ]
      }
    ]

    const showModal = true;

    const wrapper = shallowMount(Modal,{
      propsData: {
        movieTitle: movieTitle, 
        movieId: movieId, 
        cityId: cityId, 
        cityName: cityName,
        cinemas: cinemas,
      },
      data() {
        return {
          showModal,
          currentCinema: cinemas[0],
          cinemaId: cinemas[0].id,
          cinemaName: cinemas[0].name,
          seancePrice: cinemas[0].seances[0].price,
          seanceId: cinemas[0].seances[0].id,
          seanceSeats: cinemas[0].seances[0].seats,
          seanceDate: cinemas[0].seances[0].date,
          seatsIds: [1, 2]
        }
      },
    })

    expect(wrapper.find('#movie-title').text()).to.equal('Star Wars')
    expect(wrapper.find('#details').text()).to.equal('Время: 23 апреля в 15:25')
    expect(wrapper.find('#sum').text()).to.equal('Cумма: 700 руб.')
    // console.log(wrapper.html())
  })
})