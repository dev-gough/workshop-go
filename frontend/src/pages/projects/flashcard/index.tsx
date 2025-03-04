import Header from "../../../components/Header"
import { useParams } from "react-router-dom"
import { useState, useEffect } from "react"

interface Card {
    id: number
    front: string
    back: string
}

/* 

TODO: Finish porting the rating submission logic
perhaps write tests to ensure functionality is the same between the two versions?

*/

const Study = () => {
    const { deckID } = useParams<{ deckID: string }>()
    const [card, setCard] = useState<Card | null>(null)
    const [showFront, setShowFront] = useState(true)

    const fetchCard = async () => {
        try {
            const res = await fetch(`/api/flashcard/cards/${deckID}`)
            const cards = await res.json()
            const cur = Math.floor(Math.random() * cards.length)
            setCard(cards[cur])
            setShowFront(true)
        } catch(e) {
            console.error("Error fetching cards: ", e)
        }
    }

    useEffect(() => {
        fetchCard()
    }, [deckID])

    const flipCard = () => {
        setShowFront(!showFront)
    }

    return (
        <div className="bg-gray-400">
            <Header />
            <div className="lg:w-2/3 mx-auto">
                <div className="flex justify-center items-center h-screen bg-blue-100">
                    <div className="text-center">
                        {/* Flashcard Div */}
                        <div className="bg-white rounded-md shadow-md h-64 w-96 flex items-center justify-center mb-4">
                            {card && (
                                <div>{showFront ? card.front: card.back}</div>
                            )}
                        </div>
                        <button onClick={flipCard} className="text-black bg-green-400 hover:bg-green-600 px-4 py-2 rounded transition duration-300">
                            Flip Card
                        </button>
                        {/* Rating */}
                        <div className="mt-4">
                            <div className="flex justify-center items-center">
                                <label htmlFor="rating1" className="mr-2">1</label>
                                <input type="radio" id="rating1" name="rating" value="1" className="form-radio h-5 w-5 text-green-600" />
                                <label htmlFor="rating2" className="mx-2">2</label>
                                <input type="radio" id="rating2" name="rating" value="2" className="form-radio h-5 w-5 text-green-600" />
                                <label htmlFor="rating3" className="mx-2">3</label>
                                <input type="radio" id="rating3" name="rating" value="3" className="form-radio h-5 w-5 text-green-600" />
                                <label htmlFor="rating4" className="mx-2">4</label>
                                <input type="radio" id="rating4" name="rating" value="4" className="form-radio h-5 w-5 text-green-600" />
                                <label htmlFor="rating5" className="ml-2">5</label>
                                <input type="radio" id="rating5" name="rating" value="5" className="form-radio h-5 w-5 text-green-600" />
                            </div>
                        </div>
                        {/* Rating Buttons */}
                        <div className="mt-5">
                            <button id="submit-rating" className="bg-blue-400 hover:bg-blue-600 text-white px-4 py-2 rounded transition duration-300">
                                Submit Rating
                            </button>
                            <button onClick={fetchCard} id="skip-rating" className="bg-red-400 hover:bg-red-600 text-white px-4 py-2 rounded transition duration-300">
                                Skip Card
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Study